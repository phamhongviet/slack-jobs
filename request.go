package main

import (
	"net/url"
	"net/http"
)

func parseRequest(request *http.Request) (data url.Values, err error){
	data = make(url.Values)
	return data, nil
}
