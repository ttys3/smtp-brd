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
)

func init() {
	flag.String("mailgun.api_key", "", "Mailgun API key")
	flag.String("mailgun.domain", "", "Mailgun domain")
	flag.Int("mailgun.timeout", 10, "Mailgun timeout")
	registerFactory("mailgun", func() Sender {
		timeout := config.V().GetInt("mailgun.timeout")
		return NewMailgunSender(
			config.V().GetString("mailgun.domain"),
			config.V().GetString("mailgun.api_key"),
			time.Second*time.Duration(timeout),
		)
	})
}

// MailgunConfig contain settings for mailgun API
type MailgunSender struct {
	mg      *mailgun.MailgunImpl
	Domain  string
	APIKey  string
	Timeout time.Duration // TCP connection timeout
	BaseSender
}

func NewMailgunSender(domain, apiKey string, timeout time.Duration) Sender {
	if timeout == 0 {
		timeout = DefaultEmailTimeout
	}
	sender := &MailgunSender{
		Domain:  domain,
		APIKey:  apiKey,
		Timeout: timeout,
	}

	// Create an instance of the Mailgun Client
	sender.mg = mailgun.NewMailgun(domain, apiKey)
	return sender
}

func (s *MailgunSender) Name() string {
	return "mailgun"
}

func (s *MailgunSender) Send(from string, to string, subject string, bodyPlain string, bodyHtml string) error {
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
	message := s.mg.NewMessage(s.Message.From, s.Message.Subject, bodyPlain, s.Message.To...)
	message.SetHtml(bodyHtml)
	for _, to := range s.Message.CC {
		message.AddCC(to)
	}
	for _, to := range s.Message.BCC {
		message.AddBCC(to)
	}
	// extra headers used mainly for List-Unsubscribe feature
	// You can enable Mailgun’s Unsubscribe functionality by turning it on in the settings area for your domain.
	// Mailgun can automatically provide an unsubscribe footer in each email you send.
	// Mailgun will automatically prevent future emails being sent to recipients that have unsubscribed.
	// You can edit the unsubscribed address list from your Control Panel or through the API.
	// see more info via https://documentation.mailgun.com/en/latest/api-unsubscribes.html
	// and https://documentation.mailgun.com/en/latest/user_manual.html#tracking-unsubscribes
	if s.Message.Headers != nil && len(s.Message.Headers) > 0 {
		keys := make([]string, 0, len(s.Message.Headers))
		for k := range s.Message.Headers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			message.AddHeader(k, s.Message.Headers[k])
		}
	}
	s.SetTimeout(s.Timeout)
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout*2)
	defer cancel()
	// Send the message	with a 10 second timeout
	resp, id, err := s.mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("mailgun: send failed: %w", err)
	}
	fmt.Printf("mailgun: send to %s success, subject: %s, ID: %s Resp: %s\n", to, s.Message.Subject, id, resp)
	return nil
}

func (s *MailgunSender) SetTimeout(timeout time.Duration) {
	s.Timeout = timeout
	s.mg.Client().Timeout = s.Timeout
}

// String representation of Email object
func (s *MailgunSender) String() string {
	return fmt.Sprintf("provider.mailgrun: domain %s", s.Domain)
}
