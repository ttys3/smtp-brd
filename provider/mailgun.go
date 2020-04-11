// Mailgun(https://www.mailgun.com/) Free Plan provides 10,000 Emails per month

package provider

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	flag "github.com/spf13/pflag"

	"github.com/ttys3/smtp-brd/config"
	"github.com/ttys3/smtp-brd/parser"
)

func init() {
	flag.String("mg.api_key", "", "Mailgun API key")
	flag.String("mg.domain", "", "Mailgun domain")
	flag.Int("mg.timeout", 10, "Mailgun timeout")
	registerFactory("mailgun", func() Sender {
		timeout := config.V().GetInt("mg.timeout")
		return NewMailgunSender(config.V().GetString("mg.domain"), config.V().GetString("mg.api_key"), time.Second * time.Duration(timeout))
	})
}

// MailgunConfig contain settings for mailgun API
type MailgunSender struct {
	mg          *mailgun.MailgunImpl
	Domain      string
	APIKey      string
	Timeout     time.Duration // TCP connection timeout
	Message     parser.Message
	ContentType string // text/plain or text/html
}

func NewMailgunSender(Domain, APIKey string, TimeOut time.Duration) Sender {
	if TimeOut == 0 {
		TimeOut = DefaultEmailTimeout
	}
	sender := &MailgunSender {
		Domain:  Domain,
		APIKey:  APIKey,
		Timeout: TimeOut,
	}

	// Create an instance of the Mailgun Client
	sender.mg = mailgun.NewMailgun(Domain, APIKey)
	sender.mg.Client().Timeout = sender.Timeout
	return sender
}

func (s *MailgunSender) Name() string {
	return "mailgun"
}

func (s *MailgunSender) Send(from string, to string, subject string, bodyPlain string, bodyHtml string) error {
	if from != "" {
		s.Message.From = from
	}
	if subject != "" {
		s.Message.Subject = subject
	}
	message := s.mg.NewMessage(s.Message.From, s.Message.Subject, bodyPlain, to)
	message.SetHtml(bodyHtml)
	// extra headers used mainly for List-Unsubscribe feature
	// You can enable Mailgun’s Unsubscribe functionality by turning it on in the settings area for your domain.
	// Mailgun can automatically provide an unsubscribe footer in each email you send.
	// Mailgun will automatically prevent future emails being sent to recipients that have unsubscribed.
	// You can edit the unsubscribed address list from your Control Panel or through the API.
	// see more info via https://documentation.mailgun.com/en/latest/api-unsubscribes.html
	// and https://documentation.mailgun.com/en/latest/user_manual.html#tracking-unsubscribes
	if s.Message.Headers != nil && len(s.Message.Headers) > 0{
		keys := make([]string, 0, len(s.Message.Headers))
		for k := range s.Message.Headers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			message.AddHeader(k, s.Message.Headers[k])
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), DefaultEmailTimeout)
	defer cancel()
	// Send the message	with a 10 second timeout
	resp, id, err := s.mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("mailgun: send failed: %w", err)
	}
	fmt.Printf("mailgun: send to %s success, ID: %s Resp: %s\n", to, id, resp)
	return nil
}

func (s *MailgunSender) AddHeader(header, value string) {
	if s.Message.Headers == nil {
		s.Message.Headers = make(map[string]string)
	}
	s.Message.Headers[header] = value
}

func (s *MailgunSender) SetHeaders(headers map[string]string) {
	s.Message.Headers = headers
}

func (s *MailgunSender) ResetHeaders() {
	s.SetHeaders(nil)
}

func (s *MailgunSender) SetFrom(from string) {
	s.Message.From = from
}

func (s *MailgunSender) SetSubject(subject string) {
	s.Message.Subject = subject
}

func (s *MailgunSender) AddTo(to string) {
	s.Message.To = append(s.Message.To, to)
}

func (s *MailgunSender) SetTimeout(timeout time.Duration) {
	s.Timeout = timeout
	s.mg.Client().Timeout = s.Timeout
}

func (s *MailgunSender) SetCc(cc []string) {
	s.Message.Cc = cc
}

func (s *MailgunSender) SetBcc(bcc []string) {
	s.Message.Bcc = bcc
}

func (s *MailgunSender) SetDate(dt time.Time) {
	s.Message.Date = dt
}

func (s *MailgunSender) SetAttach(attach []parser.BufferAttachment) {
	s.Message.Attachments = attach
}

// String representation of Email object
func (s *MailgunSender) String() string {
	return fmt.Sprintf("provider.mailgrun: domain %s", s.Domain)
}