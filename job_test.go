package main

import (
	"net/url"
	"testing"
)

func TestNewJob(t *testing.T) {
	args := url.Values{}
	args.Set("a", "1")
	args.Set("b", "2")

	j := newJob("SlackOPS", args)

	if j.Class != "SlackOPS" {
		t.Errorf("Failed to create new job")
	}

	for _, a := range j.Args {
		if (a != "a=1") && (a != "b=2") {
			t.Errorf("Failed to create new job")
		}
	}
}
