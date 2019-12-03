package printers_test

import (
	"encoding/json"
	"testing"

	"finala/printers"
	"finala/structs"
)

type mockPrinterData struct {
	ID   int
	Name string
}

type myMockOutputMirror struct {
	mirroredOutput string
}

func (t *myMockOutputMirror) Write(p []byte) (n int, err error) {
	t.mirroredOutput += string(p)
	return len(p), nil
}

func TestGetPrice(t *testing.T) {

	mockData := []mockPrinterData{
		{ID: 1, Name: "foo"},
		{ID: 2, Name: "foo-2"},
	}

	config := []structs.PrintTableConfig{
		{Header: "ID HEADER", Key: "ID"},
		{Header: "NAME", Key: "Name"},
	}

	mockOutputMirror := &myMockOutputMirror{}

	b, _ := json.Marshal(mockData)
	printers.Table(config, b, mockOutputMirror)

	expected := `+-----------+-------+
| ID HEADER | NAME  |
+-----------+-------+
| 1         | foo   |
| 2         | foo-2 |
+-----------+-------+
`
	if mockOutputMirror.mirroredOutput != expected {
		t.Fatalf("unexpected table rendering data")
	}

}
