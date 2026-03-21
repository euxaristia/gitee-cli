package output

import (
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = origStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	r.Close()
	return string(buf[:n])
}

func TestPrintJSON(t *testing.T) {
	data := map[string]string{"key": "value"}
	out := captureStdout(t, func() {
		if err := PrintJSON(data); err != nil {
			t.Errorf("PrintJSON() error = %v", err)
		}
	})
	if !strings.Contains(out, `"key": "value"`) {
		t.Errorf("PrintJSON() output = %q, want key:value", out)
	}
}

func TestPrintJSON_Struct(t *testing.T) {
	type item struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	}
	out := captureStdout(t, func() {
		if err := PrintJSON(item{Name: "test", ID: 42}); err != nil {
			t.Errorf("PrintJSON() error = %v", err)
		}
	})
	if !strings.Contains(out, `"name": "test"`) {
		t.Errorf("PrintJSON() output = %q", out)
	}
}

func TestPrintJSON_Error(t *testing.T) {
	// Channels cannot be JSON marshaled
	ch := make(chan int)
	err := PrintJSON(ch)
	if err == nil {
		t.Error("PrintJSON(chan) expected error")
	}
}

func TestPrintTable(t *testing.T) {
	headers := []string{"NAME", "VALUE"}
	rows := [][]string{
		{"foo", "bar"},
		{"baz", "qux"},
	}
	out := captureStdout(t, func() {
		PrintTable(headers, rows)
	})
	if !strings.Contains(out, "NAME") {
		t.Errorf("PrintTable() missing headers in %q", out)
	}
	if !strings.Contains(out, "foo") {
		t.Errorf("PrintTable() missing row data in %q", out)
	}
}

func TestPrintTable_Empty(t *testing.T) {
	out := captureStdout(t, func() {
		PrintTable([]string{"A", "B"}, nil)
	})
	if !strings.Contains(out, "A") {
		t.Errorf("PrintTable() missing headers in %q", out)
	}
}

func TestFormatConstants(t *testing.T) {
	if FormatTable != "table" {
		t.Errorf("FormatTable = %q", FormatTable)
	}
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q", FormatJSON)
	}
}
