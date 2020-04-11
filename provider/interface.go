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
	SetHeader(header, value string)
	SetFrom(from string)
	AddTos(to ...string)
	AddCCs(cc ...string)
	AddBCCs(bcc ...string)
	SetSubject(subject string)
	SetTimeout(timeout time.Duration)
	SetDate(dt time.Time)
	AddAttachs(attach ...parser.BufferAttachment)
}
