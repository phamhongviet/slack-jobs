package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestParseRequest(t *testing.T) {
	body := strings.NewReader("a=1&b=2&c=3")
	request, err := http.NewRequest("POST", "http://localhost:9999/api", body)
	if err != nil {
		fmt.Println(err.Error())
	}

	data, err := parseRequest(request)

	if data.Get("a") != "1" {
		t.Errorf("Failed to parse request data: a=%s", data.Get("a"))
	}
	if data.Get("b") != "2" {
		t.Errorf("Failed to parse request data: b=%s", data.Get("a"))
	}
	if data.Get("c") != "3" {
		t.Errorf("Failed to parse request data: c=%s", data.Get("a"))
	}
}
