package http

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Request interface {
	Next() (*Response, error)
	Request() *http.Request
}

type executor struct {
	client  *Client
	request *http.Request
}

func (e *executor) Next() (*Response, error) {
	if v := e.request.Context().Value(middlewareKey); v != nil {
		if m, ok := v.(*middleware); ok {
			return m.Next()
		}
	}

	return e.doRequest()
}

func (e *executor) Request() *http.Request {
	return e.request
}

func (e *executor) call(req *http.Request) (resp *Response, err error) {
	e.request = req
	if middlewares := e.client.getMiddlewares(); len(middlewares) > 0 {
		handlers := make([]MiddlewareFunc, 0, len(middlewares)+1)
		handlers = append(handlers, e.client.getMiddlewares()...)
		handlers = append(handlers, func(r Request) (*Response, error) {
			return e.doRequest()
		})
		e.request = e.request.WithContext(context.WithValue(e.request.Context(), middlewareKey, &middleware{
			req:      e,
			handlers: handlers,
			index:    -1,
		}))
		resp, err = e.Next()
	} else {
		resp, err = e.doRequest()
	}

	return
}

// nitiate an HTTP request and return the response data.
func (e *executor) doRequest() (resp *Response, err error) {
	resp = &Response{Request: e.request}

	defer func() {
		if err != nil {
			resp = nil
		}
	}()

	for {
		resp.Response, err = e.client.Do(e.request)
		if err == nil {
			break
		}

		if resp.Response != nil {
			resp.Response.Body.Close()
		}

		if e.client.retryCount <= 0 {
			break
		}

		e.client.retryCount--

		if e.client.retryInterval > 0 {
			time.Sleep(e.client.retryInterval)
		}
	}

	return
}

func (e *executor) makeUrl(url string) string {
	if e.client.baseUrl == "" {
		return url
	}

	matched, err := regexp.MatchString(`(?i)^(http|https)://[-a-zA-Z0-9+&@#/%?=~_|,!:.;]*`, url)
	if err == nil && matched {
		return url
	}

	return strings.TrimRight(e.client.baseUrl, "/") + "/" + strings.TrimLeft(url, "/")
}
