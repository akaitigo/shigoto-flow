package sender

import (
	"context"
	"fmt"
)

type Sender interface {
	Send(ctx context.Context, to, subject, body string) error
	Type() string
}

type Config struct {
	SlackWebhookURL string
	SMTPHost        string
	SMTPPort        int
	SMTPUser        string
	SMTPPassword    string
	FromEmail       string
}

type Service struct {
	senders []Sender
}

func NewService(senders ...Sender) *Service {
	return &Service{senders: senders}
}

func (s *Service) Send(ctx context.Context, senderType, to, subject, body string) error {
	for _, sender := range s.senders {
		if sender.Type() == senderType {
			return sender.Send(ctx, to, subject, body)
		}
	}
	return fmt.Errorf("unknown sender type: %s", senderType)
}
