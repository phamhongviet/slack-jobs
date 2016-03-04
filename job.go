package main

import (
	"net/url"
)

type Job struct {
	Class string   `json:"class"`
	Args  []string `json:"args"`
}

func newJob(class string, args url.Values) Job {
	var a []string
	for k, v := range args {
		for _, vv := range v {
			a = append(a, k+"="+vv)
		}
	}

	return Job{
		Class: class,
		Args:  a,
	}
}
