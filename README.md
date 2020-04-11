# smtp-brd

## description

the main goal of this project is setup as a side container 

for remark42 comment system to send email via `WEB API`

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

## TODO

support multi `to`
add support for Cc, Bcc and attachment

## FAQ

smtp.plainAuth failed: unencrypted connection

The error is because the Go SMTP package doesn't allow authentication without encryption. 
From https://godoc.org/net/smtp#PlainAuth

    PlainAuth will only send the credentials if the connection is using TLS 
    or is connected to localhost. Otherwise authentication 
    will fail with an error, without sending the credentials.

You need either to setup TLS on your SMTP server, 
use localhost as a relay or disable authentication. 

## thanks

**mail parser**

this project use `github.com/veqryn/go-email/email` to parse email

**minimal SMTP server**

use `github.com/mhale/smtpd`, which seems based on the work of [Brad Fitzpatrick's go-smtpd](https://github.com/bradfitz/go-smtpd)

**flags parser**

`github.com/spf13/pflag`

**config manager**

`github.com/spf13/viper`

**logger**
[go.uber.org/zap](https://github.com/uber-go/zap)


