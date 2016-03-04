/*
Tiny web server to handle slack outgoing webhook and push the data to resque
*/
package main

import (
	"flag"
	"fmt"
	"github.com/fzzy/radix/redis"
	"github.com/glacjay/goini"
	"net/http"
	"strings"
)

// dictionary of string
type Dict map[string]bool

func StringToDict(s string, d Dict) {
	for _, v := range strings.Split(s, ",") {
		d[v] = true
	}
	return
}

// access rule
type AccessRule struct {
	Name      string
	Users     Dict
	Policy    bool
	Class     string
	Queue     string
	Allow_msg string
	Deny_msg  string
}

type AccessList map[string]*AccessRule

// some default global variable
var (
	PORT                   string     = ":8765"
	API_PATH               string     = "/api"
	REDIS                  string     = "localhost:6379"
	CLASS                  string     = "SlackOPS"
	QUEUE                  string     = "slackops"
	TOKENS                 Dict       = make(Dict)
	VERBOSE                bool       = false
	FLAGS                  Dict       = make(Dict)
	ACCESS_LIST            AccessList = make(AccessList)
	ALLOW_MSG                         = "please wait"
	DENY_MSG                          = "task denied"
	UNDEFINED_JOB_CAN_PASS            = true
	err                    error
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
	p_undefined_job_can_pass := flag.Bool("undefined-job-can-pass", false, "Undefined job can pass or not")
	p_allow_msg := flag.String("allow-msg", "please wait", "Specify default allow response text")
	p_deny_msg := flag.String("deny-msg", "task denied", "Specify default deny response text")
	flag.Parse()

	// mark parsed flags
	flag.Visit(check_flag)

	// read config file
	if FLAGS["C"] {
		config, err := ini.Load(*p_config)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		// load general config
		c_port, exist := config.GetString("general", "port")
		if exist {
			PORT = ":" + c_port
		}
		c_redis, exist := config.GetString("general", "redis")
		if exist {
			REDIS = c_redis
		}
		c_class, exist := config.GetString("general", "class")
		if exist {
			CLASS = c_class
		}
		c_queue, exist := config.GetString("general", "queue")
		if exist {
			QUEUE = c_queue
		}
		c_tokens, exist := config.GetString("general", "tokens")
		if exist {
			StringToDict(c_tokens, TOKENS)
		}
		c_verbose, exist := config.GetBool("general", "verbose")
		if exist {
			VERBOSE = c_verbose
		}
		c_undefined_job_can_pass, exist := config.GetBool("general", "undefined_job_can_pass")
		if exist {
			UNDEFINED_JOB_CAN_PASS = c_undefined_job_can_pass
		}
		c_allow_msg, exist := config.GetString("general", "allow_msg")
		if exist {
			ALLOW_MSG = c_allow_msg
		}
		c_deny_msg, exist := config.GetString("general", "deny_msg")
		if exist {
			DENY_MSG = c_deny_msg
		}

		// load access list
		for _, s := range config.GetSections() {
			if strings.HasPrefix(s, "job: ") {
				// read users and policy (mandatory)
				users_string, e_users := config.GetString(s, "users")
				policy_string, e_policy := config.GetString(s, "policy")
				if !(e_users && e_policy) {
					fmt.Printf("Skipping %s\nError: policy and users are mandatory in a job", s)
					continue
				}

				// read optional config
				allow_msg, _ := config.GetString(s, "allow_msg")
				deny_msg, _ := config.GetString(s, "deny_msg")
				class, _ := config.GetString(s, "class")
				queue, _ := config.GetString(s, "queue")

				// parse users and policy
				var policy bool
				if policy_string == "allow" {
					policy = true
				} else if policy_string == "deny" {
					policy = false
				} else {
					fmt.Printf("Skipping %s\nError: policy must be either 'allow' or 'deny'", s)
					continue
				}
				users := make(Dict)
				StringToDict(users_string, users)

				// create AccessRule
				ar := AccessRule{
					Name:      strings.TrimPrefix(s, "job: "),
					Users:     users,
					Policy:    policy,
					Class:     class,
					Queue:     queue,
					Allow_msg: allow_msg,
					Deny_msg:  deny_msg,
				}
				ACCESS_LIST[ar.Name] = &ar
			}
		}
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
		StringToDict(*p_tokens, TOKENS)
	}
	if FLAGS["v"] {
		VERBOSE = *p_verbose
	}
	if FLAGS["undefined-job-can-pass"] {
		UNDEFINED_JOB_CAN_PASS = *p_undefined_job_can_pass
	}
	if FLAGS["allow-msg"] {
		ALLOW_MSG = *p_allow_msg
	}
	if FLAGS["deny-msg"] {
		DENY_MSG = *p_deny_msg
	}

	if VERBOSE {
		fmt.Printf("Listening on port %s using resque at %s\n", PORT, REDIS)
		fmt.Printf("Accepting tokens:\n")
		for t, _ := range TOKENS {
			fmt.Printf("+ %s\n", t)
		}
		fmt.Printf("Jobs will be pushed to queue %s, class %s\n", QUEUE, CLASS)
	}

	// connect to redis and add queue
	rcon, err := redis.Dial("tcp", REDIS)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	rcon.Cmd("SADD", "resque:queues", QUEUE)
	// add queue optionally specified in jobs
	for k, v := range ACCESS_LIST {
		if len(v.Queue) > 0 {
			rcon.Cmd("SADD", "resque:queues", v.Queue)
			if VERBOSE {
				fmt.Printf("Job '%s' will be pushed to queue %s", k, v.Queue)
				if len(v.Class) > 0 {
					fmt.Printf(", class %s", v.Class)
				}
				fmt.Printf("\n")
			}
		}
	}
	rcon.Close()

	// start web app
	http.HandleFunc(API_PATH, apiHandler)
	err = http.ListenAndServe(PORT, nil)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
}
