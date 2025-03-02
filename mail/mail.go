package mail

import (
	"fmt"
	"strings"

	"gopkg.in/gomail.v2"

	"github.com/kurosann/aqt-core/table"
)

type Config struct {
	host string
	port int
	user string
	pwd  string
}

type Client struct {
	config Config
	dialer *gomail.Dialer
}
type Css string

func (c Css) Render(cssName string) string {
	return strings.ReplaceAll(string(c), "{{.css_name}}", cssName)
}

const (
	CssName      = "aqt-table-notify"
	TableCss Css = `<style type="text/css">
        table.{{.css_name}} {
            font-family: verdana,arial,sans-serif;
            font-size:11px;
            color:#333333;
            border-width: 1px;
            border-color: #666666;
            border-collapse: collapse;
        }
        table.{{.css_name}} th {
            border-width: 1px;
            padding: 8px;
            border-style: solid;
            border-color: #666666;
            background-color: #dedede;
			white-space:nowrap;
        }
        table.{{.css_name}} td {
            border-width: 1px;
            padding: 8px;
            border-style: solid;
            border-color: #666666;
            background-color: #ffffff;
			white-space:nowrap;
        }
    </style>`
	TableVerticalCss Css = `<style type="text/css">
    table.{{.css_name}} {
        font-family: verdana, arial, sans-serif;
        font-size: 11px;
        color: #333333;
        border-width: 1px;
        border-color: #666666;
        border-collapse: collapse;
        display: flex;
        flex-direction: column;
    }

    table.{{.css_name}} th {
        border-width: 1px;
        padding: 8px;
        border-style: solid;
        border-color: #666666;
        background-color: #dedede;
        white-space: nowrap;
    }

    table.{{.css_name}} td:first-child {
        border-width: 1px;
        padding: 8px;
        border-style: solid;
        border-color: #666666;
        background-color: #dedede;
        white-space: nowrap;
    }

    table.{{.css_name}} td {
        border-width: 1px;
        padding: 8px;
        border-style: solid;
        border-color: #666666;
        background-color: #ffffff;
        white-space: nowrap;
    }
    </style>`
)

func NewMailClient(host string, port int, user, pwd string) *Client {
	return &Client{
		config: Config{
			host: host,
			port: port,
			user: user,
			pwd:  pwd,
		},
		dialer: gomail.NewDialer(host, port, user, pwd),
	}
}

func (c Client) SendText(subject, body, to1 string, ton ...string) error {
	msg := c.genMsg(to1, subject, ton...)
	msg.SetBody("text/plain", body)
	return c.dialer.DialAndSend(msg)
}

func (c Client) SendHtml(subject, body, to1 string, ton ...string) error {
	msg := c.genMsg(subject, to1, ton...)
	msg.SetBody("text/html", body)
	return c.dialer.DialAndSend(msg)
}

func (c Client) SendTable(subject string, body any, to1 string, ton ...string) error {
	msg := c.genMsg(subject, to1, ton...)
	msg.SetBody("text/html", renderByCss(body, table.MarshalHtml, TableCss, CssName))
	return c.dialer.DialAndSend(msg)
}
func (c Client) SendVerticalTable(subject string, body any, to1 string, ton ...string) error {
	msg := c.genMsg(subject, to1, ton...)
	msg.SetBody("text/html", renderByCss(body, table.MarshalVerticalHtml, TableVerticalCss, CssName))
	return c.dialer.DialAndSend(msg)
}
func renderByCss(body any, f func(data interface{}, css ...string) []byte, css Css, cssName string) string {
	return fmt.Sprintf("%s\n%s",
		string(f(body, cssName)),
		css.Render(cssName))
}

func (c Client) genMsg(subject, to1 string, ton ...string) *gomail.Message {
	message := gomail.NewMessage()
	mailHeader := map[string][]string{
		"From":    {c.config.user},
		"To":      append([]string{to1}, ton...),
		"Subject": {subject},
	}
	message.SetHeaders(mailHeader)
	return message
}
