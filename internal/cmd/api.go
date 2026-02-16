package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/euxaristia/gitee-cli/internal/util"
)

func newAPICmd(app *App) *cobra.Command {
	var method string
	var fields []string
	var headers []string

	apiCmd := &cobra.Command{
		Use:   "api <endpoint>",
		Short: "Make a raw API request to Gitee v5",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
				return app.Client.Request(app.Ctx, strings.ToUpper(method), args[0], query, nil, nil)
			}
			var payload any
			if len(body) > 0 {
				payload = body
			}
			if err := app.Client.Request(app.Ctx, strings.ToUpper(method), args[0], query, payload, &out); err != nil {
				return err
			}
			b, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			return nil
		},
	}

	apiCmd.Flags().StringVarP(&method, "method", "X", http.MethodGet, "HTTP method")
	apiCmd.Flags().StringArrayVarP(&fields, "field", "F", nil, "Add key=value request field")
	apiCmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "Add request header key:value")
	return apiCmd
}
