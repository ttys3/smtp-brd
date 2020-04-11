module github.com/ttys3/smtp-brd

go 1.13

require (
	github.com/mailgun/mailgun-go/v4 v4.0.1
	github.com/mhale/smtpd v0.0.0-20181125220505-3c4c908952b8
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.5.0+incompatible
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.3
	github.com/veqryn/go-email v0.0.0-20200124140746-a72ac14e358c
	go.uber.org/zap v1.10.0
)

replace github.com/umputun/remark/backend => ../go/remark42/backend
