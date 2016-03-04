package main

import (
	"encoding/json"
	"github.com/fzzy/radix/redis"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Text string `json:"text"`
	}

	type Job struct {
		Class string   `json:"class"`
		Args  []string `json:"args"`
	}

	// accept only POST
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	// read and parse request body content
	body := make([]byte, r.ContentLength)
	r.Body.Read(body)
	data, err := url.ParseQuery(string(body))
	if err != nil {
		fmt.Println("error:", err)
	}

	// check token
	if !TOKENS[data.Get("token")] {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// get request and user name
	request := strings.TrimPrefix(data.Get("text"), data.Get("trigger_word"))
	request = strings.TrimLeft(request, " ")
	request = strings.TrimRight(request, " ")
	user := data.Get("user_name")

	// filter with access list
	var pass bool = UNDEFINED_JOB_CAN_PASS
	var response_text string
	if UNDEFINED_JOB_CAN_PASS {
		response_text = ALLOW_MSG
	} else {
		response_text = DENY_MSG
	}
	var class string = CLASS
	var queue string = "resque:queue:" + QUEUE

	// check if job is defined
	job_is_defined := false
	var job_type string
	for k, v := range ACCESS_LIST {
		if strings.HasPrefix(request, k) {
			if v != nil {
				job_is_defined = true
				job_type = k
				break
			}
		}
	}

	// if job is defined
	if job_is_defined {
		// if user in deny list
		if !ACCESS_LIST[job_type].Policy && ACCESS_LIST[job_type].Users[user] {
			pass = false

			// if user not in allow list
		} else if ACCESS_LIST[job_type].Policy && !ACCESS_LIST[job_type].Users[user] {
			pass = false

			// if user in allow list
		} else if ACCESS_LIST[job_type].Policy && ACCESS_LIST[job_type].Users[user] {
			pass = true

			// if user not in deny list
		} else if !ACCESS_LIST[job_type].Policy && !ACCESS_LIST[job_type].Users[user] {
			pass = true
		}

		// if user is allowed
		if pass {
			// choose response text
			if len(ACCESS_LIST[job_type].Allow_msg) > 0 {
				response_text = ACCESS_LIST[job_type].Allow_msg
			} else {
				response_text = ALLOW_MSG
			}

			// choose class
			if len(ACCESS_LIST[job_type].Class) > 0 {
				class = ACCESS_LIST[job_type].Class
			} else {
				class = CLASS
			}

			// choose queue
			if len(ACCESS_LIST[job_type].Queue) > 0 {
				queue = "resque:queue:" + ACCESS_LIST[job_type].Queue
			} else {
				queue = "resque:queue:" + QUEUE
			}

			// if user is denied
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			if len(ACCESS_LIST[job_type].Deny_msg) > 0 {
				response_text = ACCESS_LIST[job_type].Deny_msg
			} else {
				response_text = DENY_MSG
			}
		}

	} else {
		// undefined job
		if pass {
			response_text = ALLOW_MSG
			class = CLASS
			queue = "resque:queue:" + QUEUE
		} else {
			response_text = DENY_MSG
		}
	}

	if pass {
		// make connection to redis
		rcon, err := redis.Dial("tcp", REDIS)
		if err != nil {
			fmt.Println("error:", err)
		}

		// create job
		job := Job{
			Class: class,
			Args: []string{
				"request=" + request,
				"user=" + user,
				"channel_name=" + data.Get("channel_name"),
				"timestamp=" + data.Get("timestamp"),
			},
		}
		json_job, err := json.Marshal(job)
		if err != nil {
			fmt.Println("error:", err)
		}
		// push job to resque
		rcon.Cmd("RPUSH", queue, string(json_job))
		rcon.Close()
	}

	// response to slack
	res := Response{
		Text: "@" + user + ": " + response_text,
	}
	json_res, err := json.Marshal(res)
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(json_res))
}
