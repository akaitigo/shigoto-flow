package sender

import (
	"context"
	"fmt"
	"net/smtp"
)

type EmailSender struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

func NewEmail(host string, port int, user, password, from string) *EmailSender {
	return &EmailSender{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
	}
}

func (e *EmailSender) Type() string {
	return "email"
}

func (e *EmailSender) Send(_ context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", e.host, e.port)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		e.from, to, subject, body,
	)

	var auth smtp.Auth
	if e.user != "" {
		auth = smtp.PlainAuth("", e.user, e.password, e.host)
	}

	if err := smtp.SendMail(addr, auth, e.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
