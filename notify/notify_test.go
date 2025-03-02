package notify

import (
	"testing"
)

func TestSend(t *testing.T) {
	AddNotify(InitMailNotifier(
		"smtp.163.com",
		25,
		"",
		"",
	))
	AddPerson(Debug, Mail, "1343350687@qq.com")
	err := SendAllByType(Debug, "test", []map[string]interface{}{{
		"test": "test",
	}})
	if err != nil {
		t.Fatal(err.Error())
	}
}
