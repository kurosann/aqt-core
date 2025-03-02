package notify

import (
	"github.com/kurosann/aqt-core/mail"
)

type MailNotifier struct {
	client *mail.Client
}

func InitMailNotifier(host string, port int, user, pwd string) *MailNotifier {
	return &MailNotifier{
		client: mail.NewMailClient(host, port, user, pwd),
	}
}

func (m MailNotifier) Send(title string, data any, to string) error {
	return m.client.SendVerticalTable(title, data, to)
}

func (m MailNotifier) Way() Way {
	return Mail
}
