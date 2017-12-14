package main

import "net/http"
import "os"
import "fmt"
import "log"
import "html"
import "io/ioutil"

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

    if (r.Method == "GET") {
      log.Printf("GET req recd");

      w.Header().Set("Content-Type", "text/html; charset=utf-8")

      b, err := ioutil.ReadFile("test.html")

      if err != nil {
        log.Fatal(err)
      } else {
        log.Printf("%s", b)
        fmt.Fprintf(w, "%s", b)
      }
    } else if (r.Method == "POST") {
      log.Printf("POST req recd");
      fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
    }
    // fmt.Fprintf(w, "%s", r.Method)
		// fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

  log.Printf("Server started")
  log.Printf("Listening on port %s", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), nil))
}
