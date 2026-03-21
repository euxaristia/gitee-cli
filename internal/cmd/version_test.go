package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintVersionBanner(t *testing.T) {
	var buf bytes.Buffer
	printVersionBanner(&buf)
	out := buf.String()

	if !strings.Contains(out, version) {
		t.Errorf("printVersionBanner() missing version %q in output", version)
	}
	if !strings.Contains(out, "euxaristia") {
		t.Error("printVersionBanner() missing copyright")
	}
	// Check box borders
	if !strings.Contains(out, "+") {
		t.Error("printVersionBanner() missing box border")
	}
	if !strings.Contains(out, "|") {
		t.Error("printVersionBanner() missing box side")
	}
}

func TestNewVersionCmd(t *testing.T) {
	cmd := newVersionCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("version cmd error = %v", err)
	}
	if !strings.Contains(buf.String(), version) {
		t.Error("version cmd missing version string")
	}
}
