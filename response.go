/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/8/15 4:56 下午
 * @Desc: TODO
 */

package http

import (
	"github.com/dobyte/http/internal/xconv"
	"io/ioutil"
	"net/http"
	"sync"
)

type Response struct {
	*http.Response
	Request *http.Request

	err      error
	body     []byte
	bodyOnce sync.Once

	cookies     map[string]string
	cookiesOnce sync.Once

	closeErr  error
	closeOnce sync.Once
}

// ReadBody retrieves and returns the response content as []byte.
func (r *Response) ReadBody() ([]byte, error) {
	r.bodyOnce.Do(func() {
		r.body, r.err = ioutil.ReadAll(r.Response.Body)
		_ = r.Close()
	})

	return r.body[:], r.err
}

// ScanBody convert the response into a complex data structure.
func (r *Response) ScanBody(pointer interface{}) error {
	if pointer == nil {
		return nil
	}

	buf, err := r.ReadBody()
	if err != nil {
		return err
	}

	return xconv.Scan(buf, pointer)
}

// Close closes the response when it will never be used.
func (r *Response) Close() error {
	r.closeOnce.Do(func() {
		r.Response.Close = true
		r.closeErr = r.Response.Body.Close()
	})

	return r.closeErr
}

// HasHeader Determine if a header exists in the cache.
func (r *Response) HasHeader(key string) (has bool) {
	_, has = r.Header[key]
	return
}

// GetHeader Retrieve header's value from the response.
func (r *Response) GetHeader(key string) string {
	return r.Header.Get(key)
}

// GetHeaders Retrieve all header's value from the response.
func (r *Response) GetHeaders() http.Header {
	return r.Header
}

// HasCookie Determine if a cookie exists in the cache.
func (r *Response) HasCookie(key string) (has bool) {
	_, has = r.GetCookies()[key]
	return
}

// GetCookie Retrieve cookie's value from the response.
func (r *Response) GetCookie(key string) string {
	return r.GetCookies()[key]
}

// GetCookies Retrieve all cookie's value from the response.
func (r *Response) GetCookies() map[string]string {
	r.cookiesOnce.Do(func() {
		cookies := r.Cookies()
		r.cookies = make(map[string]string, len(cookies))
		for _, cookie := range cookies {
			r.cookies[cookie.Name] = cookie.Value
		}
	})

	return r.cookies
}
