package cmd

import (
	"os"
	"testing"
)

func TestPrintAny_JSON(t *testing.T) {
	// Redirect stdout
	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	err := printAny("json", nil, nil, map[string]string{"key": "value"})

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("printAny(json) error = %v", err)
	}
}

func TestPrintAny_Table(t *testing.T) {
	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	err := printAny("table", []string{"A"}, [][]string{{"1"}}, nil)

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("printAny(table) error = %v", err)
	}
}

func TestPrintAny_Empty(t *testing.T) {
	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	err := printAny("", []string{"A"}, nil, nil)

	w.Close()
	os.Stdout = origStdout
	r.Close()

	if err != nil {
		t.Errorf("printAny('') error = %v", err)
	}
}

func TestPrintAny_Unknown(t *testing.T) {
	err := printAny("xml", nil, nil, nil)
	if err == nil {
		t.Error("printAny(xml) expected error")
	}
}
