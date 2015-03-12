# Slack Jobs
A tiny web service to handle slack outgoing webhook and push the data to resque

## How it works
There are 4 components involved:
* Slack Outgoing WebHooks, https://yourteamname.slack.com/services/new
* Resque version 1, https://github.com/resque/resque/tree/1-x-stable
* Resque workers
* Slack Jobs itself

Slack Jobs run as a web service that listen to POST requests from Slack Outgoing WebHooks, filtering token and username and post jobs to resque that look like this:

```json
{
	"class": "SlackOPS",
	"args": [
		"resquest=slack do something for me",
		"user=myuser.name",
		"channel_name=slackops",
		"timestamp=1426152781.995012"
	]
}
```

## A quick test
Set up an empty redis server, for example, localhost:6379

Get Slack Jobs by running

```sh
go get github.com/phamhongviet/slack-jobs
```

Run Slack Jobs

```sh
slack-jobs -p 8765 -r localhost:6379 -undefined-job-can-pass -v -t t0k3nFromSlack0utgo1ngWebhO0ks
```

Mimic a Slack Outgoing WebHooks request

```sh
curl -X POST -d 'token=t0k3nFromSlack0utgo1ngWebhO0ks' -d 'channel_name=slackops' -d 'timestamp=1426152781.995012' -d 'user_name=myuser.name' -d 'text=ops: slack do something for me' -d 'trigger_word=ops:' localhost:8765/api
```

You should see a job in your redis server like one above.
