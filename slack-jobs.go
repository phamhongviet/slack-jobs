/*
Tiny web server to handle slack outgoing webhook and push the data to resque
*/
package main

import (
	"fmt"
	"net/http"
	"encoding/json"
)

func main() {
	http.HandleFunc("/api", apiHandler)
	http.ListenAndServe(":8765", nil)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Text string `json:"text"`
	}

	bodybytes := make([]byte, r.ContentLength)

	if r.Method == "POST" {
		r.Body.Read(bodybytes)
		res := Response{
			Text: string(bodybytes),
		}
		b, err := json.Marshal(res)
		if err != nil {
			fmt.Println("error:", err)
		}
		w.Header().Set("Content-type", "application/json")
		fmt.Fprintf(w, string(b))
	}
}
