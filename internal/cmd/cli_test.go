package cmd

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestParseMixed_FlagsAfterPositional(t *testing.T) {
	var repo string
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.StringVar(&repo, "repo", "", "")
	pos, err := parseMixed(fs, []string{"I1", "--repo", "owner/repo"})
	if err != nil {
		t.Fatalf("parseMixed error = %v", err)
	}
	if repo != "owner/repo" {
		t.Errorf("repo = %q, want owner/repo", repo)
	}
	if len(pos) != 1 || pos[0] != "I1" {
		t.Errorf("positionals = %v, want [I1]", pos)
	}
}

func TestParseMixed_ShortAndEquals(t *testing.T) {
	var method string
	var fields []string
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.StringVar(&method, "X", "GET", "")
	fs.Var((*stringsFlag)(&fields), "F", "")
	pos, err := parseMixed(fs, []string{"/ep", "-X", "POST", "-F", "a=b", "--F=c=d"})
	if err != nil {
		t.Fatalf("parseMixed error = %v", err)
	}
	if method != "POST" {
		t.Errorf("method = %q, want POST", method)
	}
	if strings.Join(fields, ",") != "a=b,c=d" {
		t.Errorf("fields = %v", fields)
	}
	if len(pos) != 1 || pos[0] != "/ep" {
		t.Errorf("positionals = %v", pos)
	}
}

func TestParseMixed_UnknownFlag(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	if _, err := parseMixed(fs, []string{"--nope"}); err == nil {
		t.Fatal("expected unknown flag error")
	}
}

func TestParseMixed_Bool(t *testing.T) {
	var private bool
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.BoolVar(&private, "private", false, "")
	pos, err := parseMixed(fs, []string{"--private", "rest"})
	if err != nil {
		t.Fatalf("parseMixed error = %v", err)
	}
	if !private {
		t.Error("private not set")
	}
	if len(pos) != 1 || pos[0] != "rest" {
		t.Errorf("positionals = %v", pos)
	}
}

func TestCommand_SetOut(t *testing.T) {
	var buf bytes.Buffer
	c := &Command{
		run: func(c *Command, args []string) error {
			c.OutOrStdout().Write([]byte(strings.Join(args, ",")))
			return nil
		},
	}
	c.SetOut(&buf)
	c.SetArgs([]string{"a", "b"})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "a,b" {
		t.Errorf("out = %q", buf.String())
	}
}
