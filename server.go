package main

import "github.com/joho/godotenv"

import "github.com/dghubble/go-twitter/twitter"
import "golang.org/x/oauth2"

import "github.com/cbroglie/mustache"

import b64 "encoding/base64"
import "encoding/json"

import "net/http"
import "os"
import "fmt"
import "log"
import "io/ioutil"
import "strings"
import "strconv"

import "time"

type BearerToken struct {
    Token_Type string
    Access_Token string
}

/*
* 1. Number of words in total
* 2. Most liked tweet
* 3. Most retweeted tweet
* 4. Total number of tweets
*
* Store in JSON and then regenerate only on request.
*/

func main() {

    TEMPLATE_FILE := "template3.html"
    // Ref time: Mon Jan 2 15:04:05 MST 2006
    begin, _ := time.Parse("2006-01-02", "2017-01-01")
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

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

        if (r.Method == "GET") {
            log.Printf("GET req recd");

            w.Header().Set("Content-Type", "text/html; charset=utf-8")

            b, err := ioutil.ReadFile("test.html")

            if err != nil {
                log.Fatal(err)
            } else {
                // log.Printf("%s", b)
                fmt.Fprintf(w, "%s", b)
            }
        } else if (r.Method == "POST") {
            r.ParseForm()

            log.Printf("POST req recd")
            w.Header().Set("Content-Type", "text/html; charset=utf-8")

            handle := r.PostForm.Get("handle")

            incRT := false

            num_tweets := 0
            word_count := 0

            var maxRT twitter.Tweet;
            var maxFav twitter.Tweet;
            firstRun := true

            last_tw_time := time.Now()
            var last_tw_id int64
            last_tw_id = 0
            var first_tw_in_period twitter.Tweet;

            log.Printf("Searching for all tweets before %s", begin)

            for last_tw_time.After(begin) {

                userTimelineParams := &twitter.UserTimelineParams{ScreenName: handle,Count: 200,IncludeRetweets: &incRT}

                if (last_tw_id > 0) {
                    userTimelineParams.MaxID = last_tw_id - 1;
                }

                tweets, _, err := client.Timelines.UserTimeline(userTimelineParams)

                if err != nil {
                    log.Fatal(err);
                } else if len(tweets) > 0 {

                    last_tw := tweets[len(tweets)-1]

                    last_tw_time, _ = time.Parse(time.RubyDate, last_tw.CreatedAt)
                    last_tw_id = last_tw.ID

                    log.Printf("Recd %d tweets; uptill %s (From ID: %d, To ID: %d); ", len(tweets), last_tw_time, tweets[0].ID, last_tw_id)

                    whole_set := last_tw_time.After(begin);

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
                            break;
                        }

                        if (!whole_set) {
                            num_tweets++;
                            first_tw_in_period = tweet;
                        }

                        word_count += len(strings.Fields(tweet.Text))

                        if (tweet.RetweetCount > maxRT.RetweetCount) {
                            maxRT = tweet
                        }

                        if (tweet.FavoriteCount > maxFav.FavoriteCount) {
                            maxFav = tweet
                        }
                    }
                }
            }

            log.Printf("Word count: %d", word_count)
            log.Printf("Number of tweets: %d", num_tweets)
            log.Printf("First tweet in this period: %d", first_tw_in_period.ID)
            log.Printf("Tweet with maximum favourites: %d", maxFav.ID)
            log.Printf("Tweet with maximum Retweets: %d", maxRT.ID)

            context := map[string]string{
                                            "num_tweets": strconv.Itoa(num_tweets),
                                            "word_count": strconv.Itoa(word_count),
                                            "handle": handle,
                                        }
            templated_res, _ := mustache.RenderFile(TEMPLATE_FILE, context)

            fmt.Fprintf(w, templated_res)
        } else {
            fmt.Fprintf(w, "Unrecognized method. Only GET / and POST / are supported!");
        }
    })

    log.Printf("Server started")
    log.Printf("Listening on port %s", os.Getenv("PORT"))
    log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), nil))
}
