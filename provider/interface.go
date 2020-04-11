package provider

import (
	"fmt"
	"time"

	"github.com/ttys3/smtp-brd/parser"
)

type Sender interface {
	fmt.Stringer
	Name() string
	// send single mail
	Send(from string, to string, subject string, bodyPlain string, bodyHtml string) error

	SetHeaders(headers map[string]string)
	ResetHeaders()
	AddHeader(header, value string)

	SetFrom(from string)
	AddTo(to string)
	SetSubject(subject string)
	SetTimeout(timeout time.Duration)
	SetCc(cc []string)
	SetBcc(cc []string)
	SetDate(dt time.Time)
	SetAttach(attach []parser.BufferAttachment)
}