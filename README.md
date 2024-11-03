# Owncast Webhook Ntfy integration

Integration between https://ntfy.sh/ and Owncast Webhook.

## What's this?

It is a small simple http server that takes a Owncast webhook and re-formats it for ntfy format.

## How to use

### Binary

[Download](https://github.com/holgerhuo/owncast-ntfy/releases/) the release binary and run it as a server

```bash
owncast-ntfy -ntfy-url "https://ntfy.sh/mytopic"

```

This will create an http (no https) server in port 8080 that will accept POST request, re-format them and send them to the ntfy url. You can use custom ntfy servers if you want.

## Options

See owncast-ntfy -h for all options.

```
Usage of owncast-ntfy:
  -allow-insecure
        Allow insecure connections to ntfy-url
  -markdown
    	Use Markdown message formatting
  -basic-auth string
        Basic auth used for ntfy, e.g.: user:pass     
  -ntfy-url string
        The ntfy url including the topic. e.g.: https://ntfy.sh/mytopic
  -port int
        The port to listen on (default 8080)   
```

# No https?

This webhook is suppose to run next to your owncast instance and only accepts local request. You should not expose this server to the internet.

# License

Originally licensed under Apache License 2.0

Now AGPL-v3.0

# Credits

Original code from [grafana-alerting-ntfy-webhook-integration](https://github.com/academo/grafana-alerting-ntfy-webhook-integration), licensed under Apache License 2.0