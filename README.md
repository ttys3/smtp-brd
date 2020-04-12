# smtp-brd

![build_container_image](https://github.com/ttys3/smtp-brd/workflows/build_container_image/badge.svg?branch=ctr)
![test_lint](https://github.com/ttys3/smtp-brd/workflows/test_lint/badge.svg?branch=master)

## description

the main goal of this project is setup as a side container

for remark42 comment system to send email via `WEB API`

## available providers

```ini
mailgun
sendgrid
```

## run with podman

```bash
sudo podman run -d --name smtpbrd -p 2525:2525 \
-e BRD_PROVIDER=sendgrid \
-e BRD_SENDGRID_API_KEY='SG-KEY-HERE' \
80x86/smtp-brd:latest
```

## run with Docker

```bash
sudo docker run -d --name smtpbrd -p 2525:2525 \
-e BRD_PROVIDER=sendgrid \
-e BRD_SENDGRID_API_KEY='SG-KEY-HERE' \
80x86/smtp-brd:latest
```

## config

there are 3 ways to config

- the config.toml file
- environment variables
- command line interface parameters

take `mailgun.api_key` for example:

config.toml

```toml
# Mailgun
[mailgun]
api_key = ""
```

the equal environment variable is `BRD_MAINGUN_API_KEY`
be aware here we introduced a prefix `BRD_` to avoid conflict

the equal cli param is `--mailgun.api_key`

run `smtp-brd --help` for full config options

the `environment variables` is recommend way for running in container.

### available config env vars and their default value

```ini
    BRD_ADDR="0.0.0.0"
    BRD_PORT="2525"
    BRD_TLS=false
    BRD_CERT="/etc/brd/ssl/ssl.crt"
    BRD_KEY="/etc/brd/ssl/ssl.key"
    BRD_DEBUG=false
    BRD_USER=""
    BRD_SECRET=""
    BRD_PROVIDER="mailgun"
    BRD_MAILGUN_API_KEY=""
    BRD_MAILGUN_DOMAIN=""
    BRD_MAILGUN_TIMEOUT=10
    BRD_SENDGRID_API_KEY=""
    BRD_SENDGRID_TIMEOUT=10
```

### system related base env vars

```ini
TZ=Asia/Hong_Kong
PUID=1000
PGID=1000
```

you can run the container in your own timezone,
just to specific the env var, for example:

```ini
TZ=America/Los_Angeles
```

## TODO

support multi `to`
add support for Cc, Bcc and attachment

## TLS

default certificate stores under `/etc/brd/ssl`

however in container, you can map it to any path as you like

just set the correct value for `BRD_CERT` and `BRD_KEY`

## FAQ

1. `smtp.plainAuth failed: unencrypted connection`

    You need either to setup TLS on your SMTP server,
    use localhost as a relay or disable authentication.
    the answer come from  
    [this issue](https://github.com/prometheus/alertmanager/issues/1358#issuecomment-386209698)

    The error is because the Go SMTP package doesn't allow authentication without encryption.
    From <https://godoc.org/net/smtp#PlainAuth>

    >   PlainAuth will only send the credentials if the connection is using TLS
        or is connected to localhost. Otherwise authentication
        will fail with an error, without sending the credentials.

2. got `x509: certificate signed by unknown authority` error while TLS enabled

    The error is because the Go net/smtp client tls.Config.InsecureSkipVerify is enforced to true

    if TLS enabled, be sure your have valid certificate

    self-signed cert is not allowed by golang smtp client

## thanks

**mail parser** this project use `github.com/veqryn/go-email/email` to parse email  
[repo](https://github.com/veqryn/go-email/email)

**minimal SMTP server** use `github.com/mhale/smtpd` [repo](https://github.com/mhale/smtpd),  
which seems based on the work of [Brad Fitzpatrick's go-smtpd](https://github.com/bradfitz/go-smtpd)

**flags parser** `github.com/spf13/pflag` [repo](https://github.com/spf13/pflag)

**config manager** `github.com/spf13/viper` [repo](https://github.com/spf13/viper)

**logger** `go.uber.org/zap` [repo](https://github.com/uber-go/zap)
