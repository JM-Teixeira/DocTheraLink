// Package mailer sends e-mail. With SMTP configured it delivers through the
// server; otherwise messages are only logged (development), so features that
// e-mail people keep working without a mail server.
package mailer

import (
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"

	"klyniva/internal/config"
)

// Message is one plain-text e-mail.
type Message struct {
	To      string
	Subject string
	Body    string
}

// Mailer sends messages.
type Mailer interface {
	Send(m Message) error
}

// New returns an SMTP mailer when SMTP_HOST is set, otherwise a LogMailer.
func New(cfg config.EmailConfig) Mailer {
	if cfg.SMTPHost == "" {
		return LogMailer{}
	}
	return &SMTPMailer{cfg: cfg}
}

// LogMailer logs that a message would have been sent (no content: it may
// be sensitive).
type LogMailer struct{}

func (LogMailer) Send(m Message) error {
	log.Printf("mailer: SMTP not configured, not sending %q to %s", m.Subject, m.To)
	return nil
}

// SMTPMailer delivers through an SMTP server (STARTTLS when offered).
type SMTPMailer struct {
	cfg config.EmailConfig
}

func (s *SMTPMailer) Send(m Message) error {
	if strings.ContainsAny(m.To+m.Subject, "\r\n") {
		return fmt.Errorf("mailer: header injection attempt")
	}
	addr := net.JoinHostPort(s.cfg.SMTPHost, s.cfg.SMTPPort)
	var auth smtp.Auth
	if s.cfg.SMTPUsername != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	}
	return smtp.SendMail(addr, auth, s.cfg.FromEmail, []string{m.To}, Format(s.cfg.FromEmail, m))
}

// Format renders the RFC 5322 message (exported for tests).
func Format(from string, m Message) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: Klyniva <%s>\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", m.To)
	fmt.Fprintf(&b, "Subject: %s\r\n", m.Subject)
	b.WriteString("MIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n")
	b.WriteString(strings.ReplaceAll(m.Body, "\n", "\r\n"))
	return []byte(b.String())
}
