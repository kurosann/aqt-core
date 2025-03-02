package table

import (
	"fmt"
	"testing"
)

type TestTable struct {
	Name   string `table:"名称"`
	Nick   string
	Ignore string `table:"-"`
	Age    int    `table:"岁"`
}

func TestMarshal(t *testing.T) {
	bs := MarshalVerticalHtml([]TestTable{{Name: "小明", Ignore: "empty", Age: 20}, {Name: "小红"}}, "test-css")
	fmt.Println(string(bs))
}
func TestMapMarshal(t *testing.T) {
	bs := MarshalHtml([]map[string]interface{}{{
		"test1": 11111111111111111,
	}, {
		"test2": "222222222222222222222221111111111111123123123123333333333333332123123333333333333333213213212311111111111111321312312131231232222222",
	}}, "test-css")
	fmt.Println(string(bs))
}
