package table

import (
	"bytes"
	"reflect"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

func MarshalHtml(data interface{}, css ...string) []byte {
	t, buffer := makeTable(genHeaderAndRows(data))
	if len(css) != 0 {
		t.Style().HTML.CSSClass = css[0]
	}
	t.RenderHTML()
	return buffer.Bytes()
}
func MarshalVerticalHtml(data interface{}, css ...string) []byte {
	header, rows := genHeaderAndRows(data)
	if len(rows) == 0 {
		return nil
	}
	target := make([]table.Row, len(rows[0]))
	for i := range target {
		target[i] = make(table.Row, len(rows)+1)
	}
	for i, h := range header {
		target[i][0] = h
	}
	for i := 0; i < len(rows); i++ {
		for j := 0; j < len(rows[i]); j++ {
			target[j][i+1] = rows[i][j]
		}
	}
	t, buffer := makeTable(nil, target)
	if len(css) != 0 {
		t.Style().HTML.CSSClass = css[0]
	}
	t.RenderHTML()
	return buffer.Bytes()
}

func Marshal(data interface{}) []byte {
	t, buffer := makeTable(genHeaderAndRows(data))
	t.Render()
	return buffer.Bytes()
}

func makeTable(header table.Row, rows []table.Row) (table.Writer, *bytes.Buffer) {
	t := table.NewWriter()
	buffer := bytes.NewBuffer([]byte{})
	t.SetOutputMirror(buffer)
	t.AppendHeader(header)
	t.AppendRows(rows)
	t.SetStyle(table.StyleLight)
	return t, buffer
}

func genHeaderAndRows(data interface{}) (table.Row, []table.Row) {
	header, rows := table.Row{}, []table.Row{}
	vOfData := reflect.ValueOf(data)
	tOfData := vOfData.Type()
	switch tOfData.Kind() {
	case reflect.Struct:
		var row table.Row
		for i := 0; i < tOfData.NumField(); i++ {
			field := tOfData.Field(i)
			val := field.Tag.Get("table")
			if val != "-" {
				columnName := ""
				if val == "" || val == "-" {
					columnName = field.Name
				} else {
					vals := strings.Split(val, ",")
					columnName = vals[0]
				}
				header = append(header, columnName)
				value := vOfData.Field(i)
				row = append(row, value.Interface())
			}
		}
		rows = append(rows, row)
	case reflect.Slice:
		for i := 0; i < vOfData.Len(); i++ {
			headers, vals := genHeaderAndRows(vOfData.Index(i).Interface())
			if len(header) == 0 {
				header = append(header, headers...)
			}
			rows = append(rows, vals...)
		}
	case reflect.Map:
		var row table.Row
		for _, key := range vOfData.MapKeys() {
			val := vOfData.MapIndex(key)
			header = append(header, key.Interface())
			row = append(row, val.Interface())
		}
		rows = append(rows, row)
	default:
		panic("table MarshalHtml not support type")
	}
	return header, rows
}
