![build_container_image](https://github.com/ttys3/smtp-brd/workflows/build_container_image/badge.svg?branch=ctr) 
![test_lint](https://github.com/ttys3/smtp-brd/workflows/test_lint/badge.svg?branch=master)

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

1. `smtp.plainAuth failed: unencrypted connection`

    The error is because the Go SMTP package doesn't allow authentication without encryption. 
    From https://godoc.org/net/smtp#PlainAuth
    
        PlainAuth will only send the credentials if the connection is using TLS 
        or is connected to localhost. Otherwise authentication 
        will fail with an error, without sending the credentials.
    
    You need either to setup TLS on your SMTP server, 
    use localhost as a relay or disable authentication. 

2. got `x509: certificate signed by unknown authority` error while TLS enabled

    The error is because the Go net/smtp client tls.Config.InsecureSkipVerify is enforced to true
    
    if TLS enabled, be sure your have valid certificate
    
    self-signed cert is not allowed by golang smtp client

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


