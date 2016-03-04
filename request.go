package main

import (
	"net/http"
	"net/url"
)

func parseRequest(request *http.Request) (data url.Values, err error) {
	body := make([]byte, request.ContentLength)
	request.Body.Read(body)
	data, err = url.ParseQuery(string(body))
	return data, err
}
