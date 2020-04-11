// SendGrid(https://sendgrid.com) Trial Plan provides 40,000 emails for 30 days
// After your trial ends, you can send 100 emails/day for free
package provider

import (
	"fmt"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	flag "github.com/spf13/pflag"

	"github.com/ttys3/smtp-brd/parser"
)

// MailgunConfig contain settings for mailgun API
type SendgridSender struct {
	sg          *sendgrid.Client
	APIKey      string        // the SendGrid API key
	Timeout     time.Duration // TCP connection timeout
	Message parser.Message
	ContentType string // text/plain or text/html
}

func init() {
	register()
}

func register() {
	flag.String("sg.api_key", "", "SendGrid API key")
	flag.Int("sg.timeout", 10, "SendGrid timeout")
	registerFactory("sendgrid", func() Sender {
		timeout, _ := flag.CommandLine.GetInt("sg.timeout")
		return NewSendgridSender(flag.Lookup("sg.api_key").Value.String(), time.Second * time.Duration(timeout))
	})
}

func NewSendgridSender(APIKey string, TimeOut time.Duration) Sender {
	if TimeOut == 0 {
		TimeOut = DefaultEmailTimeout
	}
	sender := &SendgridSender{
		APIKey:  APIKey,
		Timeout: TimeOut,
	}

	// Create an instance of the sendgrid Client
	sender.sg = sendgrid.NewSendClient(APIKey)
	return sender
}

func (s *SendgridSender) Name() string {
	return "sendgrid"
}

func (s *SendgridSender) Send(from string, to string, subject string, bodyPlain string, bodyHtml string) error {
	if from != "" {
		s.Message.From = from
	}
	if subject != "" {
		s.Message.Subject = subject
	}
	fromEmail := mail.NewEmail("", s.Message.From)
	toEmail := mail.NewEmail("", to)
	sgmail := mail.NewSingleEmail(fromEmail, s.Message.Subject, toEmail, bodyPlain, bodyHtml)

	// extra headers used mainly for List-Unsubscribe feature
	// see more info via https://sendgrid.com/docs/ui/sending-email/list-unsubscribe/
	if s.Message.Headers != nil && len(s.Message.Headers) > 0 {
		sgmail.Headers = s.Message.Headers
	}
	// Send the message	with a 10 second timeout
	sendgrid.DefaultClient.HTTPClient.Timeout = s.Timeout
	resp, err := s.sg.Send(sgmail)
	if err != nil {
		return fmt.Errorf("sendgrid: send failed: %w", err)
	}
	fmt.Printf("sendgrid: send to %s success, StatusCode: %d\n", to, resp.StatusCode)
	return nil
}

func (s *SendgridSender) AddHeader(header, value string) {
	if s.Message.Headers == nil {
		s.Message.Headers = make(map[string]string)
	}
	s.Message.Headers[header] = value
}

func (s *SendgridSender) SetHeaders(headers map[string]string) {
	s.Message.Headers = headers
}

func (s *SendgridSender) ResetHeaders() {
	s.SetHeaders(nil)
}

func (s *SendgridSender) SetFrom(from string) {
	s.Message.From = from
}

func (s *SendgridSender) SetSubject(subject string) {
	s.Message.Subject = subject
}

func (s *SendgridSender) AddTo(to string) {
	s.Message.To = append(s.Message.To, to)
}

func (s *SendgridSender) SetTimeout(timeout time.Duration) {
	s.Timeout = timeout
	sendgrid.DefaultClient.HTTPClient.Timeout = s.Timeout
}

func (s *SendgridSender) SetCc(cc []string) {
	s.Message.Cc = cc
}

func (s *SendgridSender) SetBcc(bcc []string) {
	s.Message.Bcc = bcc
}

func (s *SendgridSender) SetDate(dt time.Time) {
	s.Message.Date = dt
}

func (s *SendgridSender) SetAttach(attach []parser.BufferAttachment) {
	s.Message.Attachments = attach
}

// String representation of Email object
func (s *SendgridSender) String() string {
	return fmt.Sprintf("provider.sendgrid: API %s", "v3")
}
