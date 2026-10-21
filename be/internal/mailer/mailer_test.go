package mailer

import (
	"strings"
	"testing"

	"klyniva/internal/config"
)

func TestNewPicksImplementation(t *testing.T) {
	if _, ok := New(config.EmailConfig{}).(LogMailer); !ok {
		t.Error("without SMTP_HOST the log mailer must be used")
	}
	if _, ok := New(config.EmailConfig{SMTPHost: "smtp.example.com", SMTPPort: "587"}).(*SMTPMailer); !ok {
		t.Error("with SMTP_HOST the SMTP mailer must be used")
	}
}

func TestFormat(t *testing.T) {
	out := string(Format("noreply@klyniva.com", Message{To: "a@b.pt", Subject: "Olá", Body: "line1\nline2"}))
	for _, want := range []string{"From: Klyniva <noreply@klyniva.com>\r\n", "To: a@b.pt\r\n",
		"Subject: Olá\r\n", "charset=UTF-8", "\r\n\r\nline1\r\nline2"} {
		if !strings.Contains(out, want) {
			t.Errorf("formatted message missing %q:\n%s", want, out)
		}
	}
}

func TestRejectsHeaderInjection(t *testing.T) {
	m := &SMTPMailer{cfg: config.EmailConfig{SMTPHost: "localhost", SMTPPort: "1"}}
	if err := m.Send(Message{To: "a@b.pt\r\nBcc: evil@x.com", Subject: "x"}); err == nil ||
		!strings.Contains(err.Error(), "injection") {
		t.Errorf("want injection error, got %v", err)
	}
}
