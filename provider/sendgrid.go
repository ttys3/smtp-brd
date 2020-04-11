// SendGrid(https://sendgrid.com) Trial Plan provides 40,000 emails for 30 days
// After your trial ends, you can send 100 emails/day for free
package provider

import (
	"fmt"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	flag "github.com/spf13/pflag"

	"github.com/ttys3/smtp-brd/config"
)

// MailgunConfig contain settings for mailgun API
type SendgridSender struct {
	sg          *sendgrid.Client
	APIKey      string        // the SendGrid API key
	Timeout     time.Duration // TCP connection timeout
	BaseSender
}

func init() {
	flag.String("sendgrid.api_key", "", "SendGrid API key")
	flag.Int("sendgrid.timeout", 10, "SendGrid timeout")
	registerFactory("sendgrid", func() Sender {
		timeout := config.V().GetInt("sendgrid.timeout")
		return NewSendgridSender(config.V().GetString("sendgrid.api_key"), time.Second*time.Duration(timeout))
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
		s.SetFrom(from)
	}
	if to != "" {
		s.AddTos(to)
	}
	if subject != "" {
		s.SetSubject(subject)
	}
	// validate required param
	if s.Message.From == "" {
		return fmt.Errorf("empty From. the from object must be provided for every email send")
	}
	if len(s.Message.To) == 0 {
		return fmt.Errorf("empty to. at least one receipt should be provided")
	}

	//mail.NewV3MailInit() is quick for send single mail
	// create new *SGMailV3
	m := mail.NewV3Mail()
	// the from address must match a verified Sender Identity
	fromEmail := mail.NewEmail("", s.Message.From)
	m.SetFrom(fromEmail)
	// If present, text/plain must be first, followed by text/html
	contentPlain := mail.NewContent("text/plain", bodyPlain)
	contentHtml := mail.NewContent("text/html", bodyHtml)
	m.AddContent(contentPlain, contentHtml)

	// create new *Personalization
	personalization := mail.NewPersonalization()
	personalization.Subject = s.Message.Subject
	// Each email address in the personalization block should be unique between to, cc, and bcc
	for _, to := range s.Message.To {
		toEmail := mail.NewEmail("", to)
		personalization.AddTos(toEmail)
	}
	for _, to := range s.Message.CC {
		toEmail := mail.NewEmail("", to)
		personalization.AddCCs(toEmail)
	}
	for _, to := range s.Message.BCC {
		toEmail := mail.NewEmail("", to)
		personalization.AddBCCs(toEmail)
	}
	// extra headers used mainly for List-Unsubscribe feature
	// see more info via https://sendgrid.com/docs/ui/sending-email/list-unsubscribe/
	if s.Message.Headers != nil && len(s.Message.Headers) > 0 {
		personalization.Headers = s.Message.Headers
	}
	// add `personalization` to `m`
	m.AddPersonalizations(personalization)
	// Send the message	with a 10 second timeout
	s.SetTimeout(s.Timeout)
	resp, err := s.sg.Send(m)
	if err != nil {
		return fmt.Errorf("sendgrid: request failed: %w", err)
	}
	// 2xx responses indicate a successful request
	// see https://sendgrid.com/docs/API_Reference/Web_API_v3/Mail/errors.html
	if resp.StatusCode%100 != 2 {
		return fmt.Errorf("sendgrid: send failed with err: %+v", resp.Body)
	}
	fmt.Printf("sendgrid: send to %s success, subject: %s, StatusCode: %d\n", to, s.Message.Subject, resp.StatusCode)
	return nil
}

func (s *SendgridSender) SetTimeout(timeout time.Duration) {
	s.Timeout = timeout
	sendgrid.DefaultClient.HTTPClient.Timeout = s.Timeout
}

// String representation of Email object
func (s *SendgridSender) String() string {
	return fmt.Sprintf("provider.sendgrid: API %s", "v3")
}
