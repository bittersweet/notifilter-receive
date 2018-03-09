# Notifilter
Get notified about events

[![Build Status](https://travis-ci.org/bittersweet/notifilter-receive.svg)](https://travis-ci.org/bittersweet/notifilter-receive)

Quick explanation:

1. send data to Notifilter
2. configure what you want to be notified about and how, optionally set up rules
3. get notifications in Slack

Example:

People buy products on your website, you send all `conversion` events to Notifilter. You want to give customers that buy a certain package some special attention so you set up a notification and add a rule that the `revenue` needs to be above 100$. All incoming conversions matching the rules will be sent to a Slack channel of your choosing. You've setup the notification with a nice template so you see all relevant data right away and can click on a link to your admin page.

### Ecosystem

* [notifilter](https://github.com/bittersweet/notifilter) - Elixir/Phoenix/React app that powers the frontend
* [notifilter-rb](https://github.com/bittersweet/notifilter-rb) – Ruby gem to track events

### Architecture & Requirements

Data is received over UDP (fire and forget) and stored in Elasticsearch (for aggregation + statistics type stuff in the future). Postgres is use to store notifiers that contain notification templates (based on [Go templating](https://golang.org/pkg/html/template/)), rules and settings (send to what channel etc).

```
                persist to ES
              /
receive data
              \
                check if there are notifications
                set up that match this event      - notify channel with configured template

```

# Docker related

Build locally and add the binary to a scratch container:

```bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notifilter-receive .
docker build -t bittersweet/notifilter-receive -f Dockerfile.scratch .
```

Push to docker hub:

```
docker push bittersweet/notifilter-receive:v1.0.0
```
