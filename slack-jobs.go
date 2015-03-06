/*
Tiny web server to handle slack outgoing webhook and push the data to resque
*/
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"encoding/json"
	"strings"
	"github.com/fzzy/radix"
)

func main() {
	http.HandleFunc("/api", apiHandler)
	http.ListenAndServe(":8765", nil)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Text string `json:"text"`
	}

	body := make([]byte, r.ContentLength)

	if r.Method == "POST" {
		r.Body.Read(body)
		data, err := url.ParseQuery(string(body))
		if err != nil {
			fmt.Println("error:", err)
		}
		res := Response{
			Text: "@" + data.Get("user_name") + ": " + strings.TrimPrefix(data.Get("text"), "slackops: "),
		}
		jres, err := json.Marshal(res)
		if err != nil {
			fmt.Println("error:", err)
		}
		w.Header().Set("Content-type", "application/json")
		fmt.Fprintf(w, string(jres))
	}
}
