package provider

import (
	"time"

	"github.com/ttys3/smtp-brd/parser"
)

type BaseSender struct {
	Message     parser.Message
	ContentType string // text/plain or text/html
}

// ensure *BaseSender implement IBaseSender interface
var _ IBaseSender = &BaseSender{}

// base sender
func (s *BaseSender) SetHeader(header, value string) {
	if s.Message.Headers == nil {
		s.Message.Headers = make(map[string]string)
	}
	s.Message.Headers[header] = value
}

func (s *BaseSender) SetFrom(from string) {
	s.Message.From = from
}

func (s *BaseSender) SetSubject(subject string) {
	s.Message.Subject = subject
}

func (s *BaseSender) AddTos(to ...string) {
	s.Message.To = append(s.Message.To, to...)
}

func (s *BaseSender) AddCCs(cc ...string) {
	s.Message.CC = append(s.Message.CC, cc...)
}

func (s *BaseSender) AddBCCs(bcc ...string) {
	s.Message.BCC = append(s.Message.BCC, bcc...)
}

func (s *BaseSender) SetDate(dt time.Time) {
	s.Message.Date = dt
}

func (s *BaseSender) AddAttachs(attach ...parser.BufferAttachment) {
	s.Message.Attachments = attach
}
