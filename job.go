package main

import (
	"net/url"
)

type Job struct {
	Class string   `json:"class"`
	Args  []string `json:"args"`
}

func newJob(class string, args url.Values) Job {
	return Job{
		Class: "",
		Args:  nil,
	}
}
