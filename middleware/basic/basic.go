package basic

import (
	"encoding/base64"

	http "github.com/dobyte/http"
)

func New(config Config) http.MiddlewareFunc {
	return func(r http.Request) (*http.Response, error) {
		req := r.Request()

		if req.Header.Get("Authorization") == "" {
			auth := config.Username + ":" + config.Password
			token := base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Set("Authorization", "Basic "+token)
		}

		return r.Next()
	}
}
