package parser

import (
	"bytes"
	"time"

	"github.com/veqryn/go-email/email"
)

// we should always skip the `Received` headers
// but anyother headers may cause sensitive data leak
// so, we enforced a headers whitelist here
var headersWhitelist = []string{
	"List-Unsubscribe-Post",
	"List-Unsubscribe",
}

type BufferAttachment struct {
	Filename string
	Buffer   []byte
}

type Message struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Date        time.Time
	BodyPlain   []byte
	BodyHtml    []byte
	Attachments []BufferAttachment
	Headers     map[string]string
}

func ParseMail(message []byte) (m Message, err error) {
	r := bytes.NewReader(message)
	msg, err := email.ParseMessage(r)

	if err != nil {
		return
	}

	m.From = msg.Header.From()
	m.To = msg.Header.To()
	m.Cc = msg.Header.Cc()
	m.Bcc = msg.Header.Bcc()
	m.Subject = msg.Header.Subject()
	// @TODO does email service provider web API support this ? or they override it ?
	m.Date, _ = msg.Header.Date()

	m.Headers = make(map[string]string)
	for _, key := range headersWhitelist {
		if msg.Header.IsSet(key) {
			m.Headers[key] = msg.Header.Get(key)
		}
	}

	// single message
	if msg.HasBody() {
		m.BodyPlain = msg.Body
		m.BodyHtml = msg.Body
		return
	}

	// empty message
	if !msg.HasParts() && !msg.HasSubMessage() {
		return
	}

	// process multipart and sub message

	// image as an attachment
	// This is a multi-part message in MIME format.
	// Content-Type of "multipart"
	// Content-Type: multipart/alternative;
	// Content-Type: text/plain;
	// Content-Type: text/html;
	// Content-Type: application/octet-stream;

	// image embedded in mail body
	// Content-Type: multipart/related;
	// type="multipart/alternative";
	// Content-Transfer-Encoding: 8Bit
	// Content-Type: image/jpeg;
	//	name="306BDA5E@A6A29C69.B561915E.png.jpg"
	// Content-Transfer-Encoding: base64

	// mail reply with original mail content appendded, will have multi multipart
	// so does the same to text/html or any other mime types

	for _, part := range msg.MessagesAll() {
		mediaType, params, err := part.Header.ContentType()
		if err != nil {
			// skip error and continue
			continue
		}
		switch mediaType {
		case "text/plain":
			m.BodyPlain = append(m.BodyPlain, part.Body...)
		case "text/html":
			m.BodyHtml = append(m.BodyHtml, part.Body...)
		default:
			// we treat all other mime types as binary attachment, even image/jpeg
			if name, ok := params["name"]; ok && name != "" && len(part.Body) > 0 {
				m.Attachments = append(m.Attachments, BufferAttachment{
					Filename: name,
					Buffer:   part.Body,
				})
			}
		}
	}
	return
}