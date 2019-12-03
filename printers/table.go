package printers

import (
	"encoding/json"
	"finala/structs"
	"fmt"
	"io"
	"os"

	"github.com/jedib0t/go-pretty/table"
)

func Table(tableConfig []structs.PrintTableConfig, data []byte, out io.Writer) {

	if out == nil {
		out = os.Stdout
	}
	var dataList []map[string]interface{}
	json.Unmarshal(data, &dataList)

	t := table.NewWriter()
	t.SetOutputMirror(out)

	headerRow := table.Row{}

	for _, cell := range tableConfig {
		headerRow = append(headerRow, cell.Header)
	}

	t.AppendHeader(headerRow)
	for _, details := range dataList {
		data := []interface{}{}

		for _, row := range tableConfig {
			data = append(data, fmt.Sprintf("%v", details[row.Key]))
		}

		t.AppendRow(data)
	}

	t.Render()

}
