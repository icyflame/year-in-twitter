package main

import "github.com/joho/godotenv"

import "github.com/dghubble/go-twitter/twitter"
import "golang.org/x/oauth2"
// import "github.com/dghubble/oauth1"

import b64 "encoding/base64"
import "encoding/json"

import "net/http"
import "os"
import "fmt"
import "log"
import "io/ioutil"
import "strings"

type BearerToken struct {
    Token_Type string
    Access_Token string
}

func main() {
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

        log.Printf("%s", b64_token)

        req.Header.Add("Authorization", "Basic " + b64_token)
        req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

        access_tok_client := &http.Client{}

        resp, err := access_tok_client.Do(req)

        if err != nil {
            log.Fatal(err)
        } else {
            log.Printf("%s", resp)
            bearer, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                log.Fatal(err)
            } else {
                log.Printf("Bearer token: %s", bearer)

                err := json.Unmarshal(bearer, &tok)

                if err != nil {
                    log.Fatal(err)
                } else {
                    log.Printf("%+v", tok)
                    log.Printf("%s", tok.Access_Token)
                }
            }
        }
    }

    log.Printf("Access token: %s", tok.Access_Token);

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
            log.Print(r.PostForm)

            handle := r.PostForm.Get("handle")
            fmt.Fprintf(w, "Hello, %s", handle)

            incRT := false

            userTimelineParams := &twitter.UserTimelineParams{ScreenName: handle, Count: 2, IncludeRetweets: &incRT}
            tweets, resp, err := client.Timelines.UserTimeline(userTimelineParams)
            log.Printf("USER TIMELINE: %+v", tweets)
            log.Printf("RESP: %+v", resp);
            log.Print("Error:");
            log.Printf("%+v", err);
        } else {
            fmt.Fprintf(w, "Unrecognized method. Only GET / and POST / are supported!");
        }
    })

    log.Printf("Server started")
    log.Printf("Listening on port %s", os.Getenv("PORT"))
    log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), nil))
}
