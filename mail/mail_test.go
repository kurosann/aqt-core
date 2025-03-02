package mail

import (
	"testing"

	"gopkg.in/gomail.v2"
)

func TestMail(t *testing.T) {
	message := gomail.NewMessage()
	message.SetHeader("From", "kuro143@163.com")
	message.SetHeader("To", "1343350687@qq.com")
	message.SetHeader("Subject", "gotest")
	message.SetBody("text/plain", "a go test body")
	dialer := gomail.NewDialer("smtp.163.com", 465, "kuro143@163.com", "FKBUBELQOFGWFIRH")
	err := dialer.DialAndSend(message)
	if err != nil {
		t.Fatal(err.Error())
	}
}

type TestTable struct {
	Name   string `table:"名称"`
	Nick   string
	Ignore string `table:"-"`
	Age    int    `table:"岁"`
}

func TestExample(t *testing.T) {
	client := NewMailClient("smtp.163.com", 465, "kuro143@163.com", "FKBUBELQOFGWFIRH")
	//err := client.SendTable("TestSendTable", []TestTable{{Name: "小明", Ignore: "empty", Age: 20}, {Name: "小红"}}, "1343350687@qq.com")
	//if err != nil {
	//	t.Fatalf(err.Error())
	//}
	err := client.SendVerticalTable("TestSendTable", []TestTable{{Name: "小明", Ignore: "empty", Age: 20}, {Name: "小红"}}, "1343350687@qq.com")
	if err != nil {
		t.Fatal(err.Error())
	}
}
