/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/8/16 9:47 上午
 * @Desc: request's middleware
 */

package http

type MiddlewareFunc = func(r Request) (*Response, error)

const middlewareKey = "__httpClientMiddlewareKey"

type middleware struct {
	err      error
	req      Request
	resp     *Response
	index    int
	handlers []MiddlewareFunc
}

// Next exec the next middleware.
func (m *middleware) Next() (*Response, error) {
	if m.index < len(m.handlers) {
		m.index++
		if m.resp, m.err = m.handlers[m.index](m.req); m.err != nil {
			return m.resp, m.err
		}
	}

	return m.resp, m.err
}
