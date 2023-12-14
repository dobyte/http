/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/8/16 9:40 上午
 * @Desc: TODO
 */

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"github.com/dobyte/http/internal"
	"net/http"
	"regexp"
	"strings"
)

const (
	MethodGet     = http.MethodGet
	MethodHead    = http.MethodHead
	MethodPost    = http.MethodPost
	MethodPut     = http.MethodPut
	MethodPatch   = http.MethodPatch
	MethodDelete  = http.MethodDelete
	MethodConnect = http.MethodConnect
	MethodOptions = http.MethodOptions
	MethodTrace   = http.MethodTrace
)

type request struct {
	executor
}

type RequestOptions struct {
	Headers map[string]string
	Cookies map[string]string
}

func newRequest(client *Client) *request {
	return &request{executor{client: client}}
}

// send a http request.
func (r *request) request(method, url string, data interface{}, opts ...*RequestOptions) (*Response, error) {
	req, err := r.prepare(method, url, data, opts...)
	if err != nil {
		return nil, err
	}

	return r.call(req)
}

// build a http request.
func (r *request) prepare(method, url string, data interface{}, opts ...*RequestOptions) (req *http.Request, err error) {
	var (
		buf     []byte
		body    = bytes.NewBuffer(nil)
		headers = r.client.GetHeaders()
		cookies = r.client.GetCookies()
	)

	method, url = strings.ToUpper(method), r.makeUrl(url)

	if len(opts) > 0 && opts[0] != nil {
		for key, value := range opts[0].Headers {
			headers[key] = value
		}

		for key, value := range opts[0].Cookies {
			cookies[key] = value
		}
	}

	switch contentType := headers[HeaderContentType]; contentType {
	case ContentTypeJson, ContentTypeXml, ContentTypeFormUrlEncoded:
		switch v := data.(type) {
		case nil:
			// ignore
		case string:
			buf = []byte(v)
		case []byte:
			buf = v[:]
		default:
			switch contentType {
			case ContentTypeJson:
				buf, err = json.Marshal(data)
				if err != nil {
					return
				}
			case ContentTypeXml:
				buf, err = xml.Marshal(data)
				if err != nil {
					return
				}
			case ContentTypeFormUrlEncoded:
				buf = []byte(internal.BuildParams(data))
			}
		}

		body.Write(buf)
	default:
		switch v := data.(type) {
		case nil:
			// ignore
		case string:
			buf = []byte(v)
		case []byte:
			buf = v
		default:
			switch method {
			case MethodGet, MethodPost:
				buf = []byte(internal.BuildParams(data))
			}
		}

		if len(buf) > 0 {
			if (buf[0] == '[' || buf[0] == '{') && json.Valid(buf) {
				headers[HeaderContentType] = ContentTypeJson
				body.Write(buf)
			} else if matched, _ := regexp.Match(`^[\w\[\]]+=.+`, buf); matched {
				if method != MethodGet {
					headers[HeaderContentType] = ContentTypeFormUrlEncoded
					body.Write(buf)
				}
			} else {
				body.Write(buf)
			}
		}
	}

	if method == MethodGet {
		if _, ok := headers[HeaderContentType]; !ok && len(buf) > 0 {
			if strings.Contains(url, "?") {
				url = url + "&" + string(buf)
			} else {
				url = url + "?" + string(buf)
			}
		}
	}

	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return
	}

	if r.client.ctx != nil {
		req = req.WithContext(r.client.ctx)
	} else {
		req = req.WithContext(context.Background())
	}

	for key, value := range headers {
		switch key {
		case HeaderCookie:
			// ignore
		default:
			req.Header.Set(key, value)
		}
	}

	if len(cookies) > 0 {
		slice := make([]string, 0, len(cookies))
		for key, value := range r.client.cookies {
			slice = append(slice, key+"="+value)
		}
		req.Header.Set(HeaderCookie, strings.Join(slice, ";"))
	}

	if host := req.Header.Get(HeaderHost); host != "" {
		req.Host = host
	}

	return
}
