package ctxclient

import (
	"fmt"
	"strings"
)

type (
	QSParams map[string]string
)

func addQSParams(url string, params QSParams) string {
	if len(params) > 0 {
		l := []string{}
		url += "?"
		for key, val := range params {
			l = append(l, fmt.Sprintf("%s=%s", key, val))
		}
		qs := strings.Join(l, "&")
		url += qs
	}
	return url
}
