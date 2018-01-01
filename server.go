package main

import "github.com/joho/godotenv"

import "github.com/dghubble/go-twitter/twitter"
import "golang.org/x/oauth2"

import "github.com/go-redis/redis"

import b64 "encoding/base64"
import "encoding/json"

import "net/http"
import "os"
import "fmt"
import "log"
import "io/ioutil"
import "strings"

import "html/template"

import "bytes"

import "time"

type BearerToken struct {
    Token_Type string
    Access_Token string
}

type OEmbedWithId struct {
    ID int64
    Tweet *twitter.OEmbedTweet
}

type TemplateContext struct {
    Num_Tweets int
    Word_Count int
    Handle string
    Most_Fav template.HTML
    Most_RT template.HTML
    First_Tweet template.HTML
    Last_Tweet template.HTML
    Most_Fav_Count int
    Most_RT_Count int
    MonthNames []string
    MonthValues []int
    WeekdayNames []string
    WeekdayValues []int
}

func StringifyContext (a TemplateContext) string {
    return fmt.Sprintf("%v", a);
}

func UnstringifyContext (a string) TemplateContext {
    var res TemplateContext
    fmt.Sscanf(a, "%v", &res)
    return res
}

func GetOEmbedTw(tw int64, tw_chans chan OEmbedWithId, client *twitter.Client) {
    log.Printf("%d", tw)
    statusOembedParams := &twitter.StatusOEmbedParams{ID: tw, MaxWidth: 500}
    oembed, _, _ := client.Statuses.OEmbed(statusOembedParams)
    var oembed_id OEmbedWithId
    oembed_id.ID = tw
    oembed_id.Tweet = oembed
    tw_chans <- oembed_id
}

/*
* 1. Number of words in total
* 2. Most liked tweet
* 3. Most retweeted tweet
* 4. Total number of tweets
* 5. Pie chart for tweets by month and tweets by weekday
*
* Store in JSON and then regenerate only on request.
*/

func main() {

    months := []string{
                        "January",
                        "February",
                        "March",
                        "April",
                        "May",
                        "June",
                        "July",
                        "August",
                        "September",
                        "October",
                        "November",
                        "December",
                    }

    weekdays := []string{
                            "Monday",
                            "Tuesday",
                            "Wednesday",
                            "Thursday",
                            "Friday",
                            "Saturday",
                            "Sunday",
                        }

    log.Printf("Months: %s", strings.Join(months, ","))

    TEMPLATE_FILE := "template.html"
    // Ref time: Mon Jan 2 15:04:05 MST 2006
    begin, _ := time.Parse("2006-01-02", "2017-01-01")
    end, _ := time.Parse("2006-01-02", "2018-01-01")
    // begin, _ := time.Parse("2006-01-02", "2017-12-01")

    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    var tok BearerToken

    req, err := http.NewRequest("POST", "https://api.twitter.com/oauth2/token", strings.NewReader("grant_type=client_credentials"))

    if err != nil {
        log.Fatal(err)
    } else {

        data := os.Getenv("CONSUMER_KEY") + ":" + os.Getenv("CONSUMER_SECRET")
        b64_token := b64.StdEncoding.EncodeToString([]byte(data))

        req.Header.Add("Authorization", "Basic " + b64_token)
        req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

        access_tok_client := &http.Client{}

        resp, err := access_tok_client.Do(req)

        if err != nil {
            log.Fatal(err)
        } else {
            bearer, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                log.Fatal(err)
            } else {
                err := json.Unmarshal(bearer, &tok)

                if err != nil {
                    log.Fatal(err)
                } else {
                    log.Printf("Access token received")
                }
            }
        }
    }

    config := &oauth2.Config{}
    token := &oauth2.Token{AccessToken: tok.Access_Token}
    httpClient := config.Client(oauth2.NoContext, token)
    client := twitter.NewClient(httpClient)

	redClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_SERVER") + ":" + os.Getenv("REDIS_PORT"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := redClient.Ping().Result()

    log.Printf("Redis client connected. Ping response: %v", pong)
    redClientExists := err == nil

    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

        if (r.Method == "GET" && r.URL.Path == "/") {
            log.Printf("GET req recd");

            w.Header().Set("Content-Type", "text/html; charset=utf-8")

            b, err := ioutil.ReadFile("index.html")

            if err != nil {
                log.Fatal(err)
                fmt.Fprintf(w, "There has been an error! Error: %v", err)
                return;
            } else {
                fmt.Fprintf(w, "%s", b)
                return;
            }
        }

        if (r.Method == "POST" && r.URL.Path == "/") {
            r.ParseForm()
            handle := r.PostForm.Get("handle")
            http.Redirect(w, r, "/get/" + handle, 302)
            return;
        }

        if (r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/get/")) {
            log.Println("GET ", r.URL.Path)
            w.Header().Set("Content-Type", "text/html; charset=utf-8")

            handle := strings.Replace(r.URL.Path, "/get/", "", 1)

            // TODO: Check if this username's data is already there in redis
            if (redClientExists) {
                log.Printf("Check if data is in Redis, if it is there then don't do any of the stuff below!")
                val, err := redClient.Get(handle).Result()
                log.Println("Value: ", val)
                log.Println("Error: ", err)
                if len(val) > 0 {
                    res := UnstringifyContext(val)
                    log.Printf("Retrieved from redis for %s: %v", handle, res)
                }
            }

            incRT := false

            num_tweets := 0
            word_count := 0

            var maxRT twitter.Tweet;
            var maxFav twitter.Tweet;
            firstRun := true

            last_tw_time := time.Now()
            var last_tw_id int64
            last_tw_id = 0
            var first_tw_in_period, last_tw_in_period twitter.Tweet;
            lastTweetFound := false

            monthMap := map[string]int{}
            weekdayMap := map[string]int{}
            hourMap := map[int]int{}

            log.Printf("Searching for all tweets before %s", begin)

            for last_tw_time.After(begin) {

                userTimelineParams := &twitter.UserTimelineParams{ScreenName: handle,Count: 200,IncludeRetweets: &incRT}

                if (last_tw_id > 0) {
                    userTimelineParams.MaxID = last_tw_id - 1;
                }

                tweets, _, err := client.Timelines.UserTimeline(userTimelineParams)

                if err != nil {
                    fmt.Fprintf(w, "Error: %v", err)
                    return;
                } else if len(tweets) > 0 {

                    first_tw := tweets[0]
                    first_tw_time, _ := time.Parse(time.RubyDate, first_tw.CreatedAt)

                    last_tw := tweets[len(tweets)-1]

                    last_tw_time, _ = time.Parse(time.RubyDate, last_tw.CreatedAt)
                    last_tw_id = last_tw.ID

                    log.Printf("Recd %d tweets; uptill %s (From ID: %d, To ID: %d); ", len(tweets), last_tw_time, tweets[0].ID, last_tw_id)

                    whole_set := last_tw_time.After(begin) && first_tw_time.Before(end)

                    if (whole_set) {
                        num_tweets += len(tweets)
                        first_tw_in_period = last_tw;
                    }

                    if firstRun {
                        maxRT = tweets[0]
                        maxFav = tweets[0]
                        firstRun = false
                    }

                    for _, tweet := range tweets {
                        this_tw_time, _ := time.Parse(time.RubyDate, tweet.CreatedAt)
                        if !whole_set && this_tw_time.Before(begin) {
                            // If this tweet is before the beginning of this
                            // period, then every tweet after WILL definitely be
                            // before this period and not of any interest to us.
                            break;
                        }

                        if (!whole_set && this_tw_time.After(end)) {
                            // If this tweet is after the end of the period,
                            // then we should skip this tweet and keep checking
                            // older tweets
                            continue;
                        }

                        if (!lastTweetFound) {
                            last_tw_in_period = tweet
                            lastTweetFound = true
                        }

                        if (!whole_set) {
                            num_tweets++
                            first_tw_in_period = tweet
                        }

                        word_count += len(strings.Fields(tweet.Text))

                        if (tweet.RetweetCount > maxRT.RetweetCount) {
                            maxRT = tweet
                        }

                        if (tweet.FavoriteCount > maxFav.FavoriteCount) {
                            maxFav = tweet
                        }

                        monthMap[this_tw_time.Month().String()] += 1
                        weekdayMap[this_tw_time.Weekday().String()] += 1
                        hourMap[this_tw_time.Hour()] += 1
                    }
                }
            }

            log.Printf("Word count: %d", word_count)
            log.Printf("Number of tweets: %d", num_tweets)
            log.Printf("First tweet in this period: %d", first_tw_in_period.ID)
            log.Printf("Tweet with maximum favourites: %d", maxFav.ID)
            log.Printf("Tweet with maximum Retweets: %d", maxRT.ID)
            log.Printf("Month map: %+v", monthMap)
            log.Printf("Weekday map: %+v", weekdayMap)
            log.Printf("Hour map: %+v", hourMap)

            id_list := []int64{ first_tw_in_period.ID, maxFav.ID, maxRT.ID, last_tw_in_period.ID }

            tweet_chans := make(chan OEmbedWithId)

            for _, tw_id := range id_list {
                go GetOEmbedTw(tw_id, tweet_chans, client)
            }

            tweet1 := <-tweet_chans
            tweet2 := <-tweet_chans
            tweet3 := <-tweet_chans
            tweet4 := <-tweet_chans

            tw_coll := []OEmbedWithId{tweet1, tweet2, tweet3, tweet4}
            reqd_tweets := map[int64]*twitter.OEmbedTweet{}

            for _, t := range tw_coll {
                reqd_tweets[t.ID] = t.Tweet
            }

            ftw := reqd_tweets[first_tw_in_period.ID]
            mft := reqd_tweets[maxFav.ID]
            mrt := reqd_tweets[maxRT.ID]
            ltw := reqd_tweets[last_tw_in_period.ID]

            new_temp, err := template.ParseFiles(TEMPLATE_FILE)

            monthValues := make([]int, 12)

            for ind, month := range months {
                monthValues[ind] = monthMap[month]
            }

            weekdayValues := make([]int, 7)
            for ind, weekday := range weekdays {
                weekdayValues[ind] = weekdayMap[weekday]
            }

            if err != nil {
                log.Fatal(err)
            } else {
                data_obj := TemplateContext{
                    num_tweets,
                    word_count,
                    handle,
                    template.HTML(mft.HTML),
                    template.HTML(mrt.HTML),
                    template.HTML(ftw.HTML),
                    template.HTML(ltw.HTML),
                    maxFav.FavoriteCount,
                    maxRT.RetweetCount,
                    months,
                    monthValues,
                    weekdays,
                    weekdayValues,
                }

                var templated_res bytes.Buffer
                new_temp.Execute(&templated_res, data_obj)

                fmt.Fprint(w, templated_res.String())
                if (redClientExists) {
                    err := redClient.Set(handle, StringifyContext(data_obj), 0).Err()
                    if err != nil {
                        log.Printf("Couldn't write %s's data to Redis: %v", handle, err)
                    } else {
                        log.Printf("Wrote %s's data to Redis", handle)
                    }
                }
                return;
            }
        }

        fmt.Fprintf(w, r.URL.Path + " is not supported! Only GET / and POST / is supported right now.");
        return;
    })

    log.Printf("Server started")
    log.Printf("Listening on port %s", os.Getenv("PORT"))
    log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), nil))
}
