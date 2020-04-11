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
