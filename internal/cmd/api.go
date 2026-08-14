package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/euxaristia/gitee-cli/internal/util"
)

func newAPICmd(app *App) *Command {
	return &Command{
		Use:   "api <endpoint>",
		Short: "Make a raw API request to Gitee v5",
		run: func(c *Command, args []string) error {
			return runAPI(app, args)
		},
	}
}

func runAPI(app *App, args []string) error {
	method := http.MethodGet
	var fields []string
	var headers []string
	pos, err := parseArgs("api", args, func(fs *flag.FlagSet) {
		fs.StringVar(&method, "method", http.MethodGet, "HTTP method")
		fs.StringVar(&method, "X", http.MethodGet, "HTTP method")
		fs.Var((*stringsFlag)(&fields), "field", "Add key=value request field")
		fs.Var((*stringsFlag)(&fields), "F", "Add key=value request field")
		fs.Var((*stringsFlag)(&headers), "header", "Add request header key:value")
		fs.Var((*stringsFlag)(&headers), "H", "Add request header key:value")
	})
	if err != nil {
		return err
	}
	if err := exactArgs(pos, 1); err != nil {
		return err
	}
	if !strings.EqualFold(method, http.MethodGet) {
		if err := ensureToken(app); err != nil {
			return err
		}
	}
	query := map[string]string{}
	body := map[string]any{}
	for _, field := range fields {
		k, v, err := util.KeyValue(field)
		if err != nil {
			return err
		}
		if strings.EqualFold(method, http.MethodGet) {
			query[k] = v
		} else {
			body[k] = v
		}
	}
	_ = headers
	var out any
	if strings.EqualFold(method, http.MethodDelete) {
		return app.Client.Request(app.Ctx, strings.ToUpper(method), pos[0], query, nil, nil)
	}
	var payload any
	if len(body) > 0 {
		payload = body
	}
	if err := app.Client.Request(app.Ctx, strings.ToUpper(method), pos[0], query, payload, &out); err != nil {
		return err
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
