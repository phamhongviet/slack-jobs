package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	URL := "http://localhost:8765/api"
	token := "token1asdfQWER"
	timeout := 1 * time.Second

	data := url.Values{}
	data.Set("token", token)
	data.Set("channel_name", "slackops")
	data.Set("timestamp", "1426152781.995012")
	data.Set("user_name", "steve.jobs")
	data.Set("text", "ops: make toys ipod")
	data.Set("trigger_word", "ops:")

	client := &http.Client{
		Timeout: timeout,
	}

	response, err := client.PostForm(URL, data)

	if err != nil {
		t.Errorf(err.Error())
	}

	body, _ := ioutil.ReadAll(response.Body)
	bodyString := string(body[:])
	if bodyString != `{"text":"@steve.jobs: OK, one sec"}` {
		t.Errorf(bodyString)
	}
}