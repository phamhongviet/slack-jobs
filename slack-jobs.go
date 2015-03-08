/*
Tiny web server to handle slack outgoing webhook and push the data to resque
*/
package main

import (
	"fmt"
	"flag"
	"net/http"
	"net/url"
	"encoding/json"
	"strings"
	"github.com/fzzy/radix/redis"
	"github.com/glacjay/goini"
)

// some default global variable
var (
	PORT string = "8765"
	API_PATH string = "/api"
	REDIS string = "localhost:6379"
	CLASS string = "SlackOPS"
	QUEUE string = "slackops"
	TOKENS map[string]bool = make(map[string]bool)
	VERBOSE bool = false
	FLAGS map[string]bool = make(map[string]bool)
	err error
)

func check_flag(f *flag.Flag) {
	FLAGS[f.Name] = true
}

func main() {
	// config parameter
	p_port := flag.String("p", "8765", "listen port")
	p_redis := flag.String("r", "localhost:6379", "redis host")
	p_class := flag.String("c", "SlackOPS", "resque class")
	p_queue := flag.String("q", "slackops", "resque queue")
	p_tokens := flag.String("t", "", "slack tokens, split by a comma (,)")
	p_verbose := flag.Bool("v", false, "verbose")
	p_config := flag.String("C", "", "configuration file")
	flag.Parse()

	// mark parsed flags
	flag.Visit(check_flag)
	fmt.Printf("flags: %s\n", FLAGS)

	// read config file
	if len(*p_config) > 0 {
		config, err := ini.Load(*p_config)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		c_port, exist := config.GetString("general", "port")
		if exist {
			PORT = ":" + c_port
		}
		// TODO: other config
	}

	// override configuration with flags
	if FLAGS["p"] {
		PORT = ":" + *p_port
	}
	if FLAGS["r"] {
		REDIS = *p_redis
	}
	if FLAGS["c"] {
		CLASS = *p_class
	}
	if FLAGS["q"] {
		QUEUE = *p_queue
	}
	if FLAGS["t"] {
		for _, t := range strings.Split(*p_tokens, ",") {
			TOKENS[t] = true
		}
	}
	if FLAGS["v"] {
		VERBOSE = *p_verbose
	}

	if VERBOSE {
		fmt.Printf("Listening on port %s\n", PORT)
	}

	// connect to redis and add queue
	rcon, err := redis.Dial("tcp", REDIS)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	rcon.Cmd("SADD", "resque:queues", QUEUE)
	rcon.Close()

	// start web app
	http.HandleFunc(API_PATH, apiHandler)
	http.ListenAndServe(PORT, nil)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Text string `json:"text"`
	}

	type Job struct {
		Class string `json:"class"`
		Args map[string]string `json:"args"`
	}

	// accept only POST
	if r.Method != "POST" {
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
	if ! TOKENS[data.Get("token")] {
		return
	}

	// make connection to redis
	rcon, err := redis.Dial("tcp", REDIS)
	if err != nil {
		fmt.Println("error:", err)
	}

	// create job
	job := Job{
		Class: CLASS,
		Args: map[string]string {
		"request": strings.TrimPrefix(data.Get("text"), data.Get("trigger_word")),
		"user": data.Get("user_name"),
		"channel": data.Get("channel_name"),
		"timestamp": data.Get("timestamp"),
		},
	}
	jjob, err := json.Marshal(job)
	if err != nil {
		fmt.Println("error:", err)
	}
	// push job to resque
	queue := "resque:queue:" + QUEUE
	rcon.Cmd("RPUSH", queue, string(jjob))
	rcon.Close()

	// response to slack
	res := Response{
		Text: "@" + data.Get("user_name") + ": " + strings.TrimPrefix(data.Get("text"), "slackops: "),
	}
	jres, err := json.Marshal(res)
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Header().Set("Content-type", "application/json")
	fmt.Fprintf(w, string(jres))
}
