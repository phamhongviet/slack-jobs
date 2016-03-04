package main

import (
	"net/url"
)

type job struct {
	Class string   `json:"class"`
	Args  []string `json:"args"`
}

func newJob(class string, args url.Values) job {
	var a []string
	for k, v := range args {
		for _, vv := range v {
			a = append(a, k+"="+vv)
		}
	}

	return job{
		Class: class,
		Args:  a,
	}
}
