package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Command is a stdlib command. Tests set args and IO, then call Execute.
type Command struct {
	Use          string
	Short        string
	SilenceUsage bool
	args         []string
	stdout       io.Writer
	stderr       io.Writer
	stdin        io.Reader
	run          func(*Command, []string) error
}

func (c *Command) SetArgs(args []string) { c.args = args }
func (c *Command) SetOut(w io.Writer)    { c.stdout = w }
func (c *Command) SetErr(w io.Writer)    { c.stderr = w }
func (c *Command) SetIn(r io.Reader)     { c.stdin = r }

func (c *Command) OutOrStdout() io.Writer {
	if c != nil && c.stdout != nil {
		return c.stdout
	}
	return os.Stdout
}

func (c *Command) ErrOrStderr() io.Writer {
	if c != nil && c.stderr != nil {
		return c.stderr
	}
	return os.Stderr
}

func (c *Command) InOrStdin() io.Reader {
	if c != nil && c.stdin != nil {
		return c.stdin
	}
	return os.Stdin
}

func (c *Command) Execute() error {
	if c.run == nil {
		return nil
	}
	args := c.args
	if args == nil {
		args = []string{}
	}
	return c.run(c, args)
}

type stringsFlag []string

func (s *stringsFlag) String() string { return strings.Join(*s, ",") }

func (s *stringsFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func parseArgs(name string, args []string, setup func(*flag.FlagSet)) ([]string, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	setup(fs)
	return parseMixed(fs, args)
}

func parseMixed(fs *flag.FlagSet, args []string) ([]string, error) {
	var positionals []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		name, value, hasValue, isFlag := splitFlag(a)
		if !isFlag {
			positionals = append(positionals, a)
			continue
		}
		if name == "help" || name == "h" {
			return nil, flag.ErrHelp
		}
		f := fs.Lookup(name)
		if f == nil {
			return nil, fmt.Errorf("flag provided but not defined: -%s", name)
		}
		if hasValue {
			if err := f.Value.Set(value); err != nil {
				return nil, err
			}
			continue
		}
		if isBoolFlag(f) {
			if err := f.Value.Set("true"); err != nil {
				return nil, err
			}
			continue
		}
		if i+1 >= len(args) {
			return nil, fmt.Errorf("flag needs an argument: -%s", name)
		}
		i++
		if err := f.Value.Set(args[i]); err != nil {
			return nil, err
		}
	}
	return positionals, nil
}

func splitFlag(arg string) (name, value string, hasValue, ok bool) {
	if arg == "-" || !strings.HasPrefix(arg, "-") {
		return "", "", false, false
	}
	body := strings.TrimPrefix(arg, "-")
	body = strings.TrimPrefix(body, "-")
	if i := strings.IndexByte(body, '='); i >= 0 {
		return body[:i], body[i+1:], true, true
	}
	return body, "", false, true
}

func isBoolFlag(f *flag.Flag) bool {
	bf, ok := f.Value.(interface{ IsBoolFlag() bool })
	return ok && bf.IsBoolFlag()
}

func exactArgs(args []string, n int) error {
	if len(args) != n {
		return fmt.Errorf("accepts %d arg(s), received %d", n, len(args))
	}
	return nil
}

func maxArgs(args []string, n int) error {
	if len(args) > n {
		return fmt.Errorf("accepts at most %d arg(s), received %d", n, len(args))
	}
	return nil
}

func requireFlag(name, value string) error {
	if value == "" {
		return fmt.Errorf("required flag --%s not set", name)
	}
	return nil
}
