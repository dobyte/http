/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/8/14 4:11 下午
 * @Desc: TODO
 */

package http

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"
)

type Client struct {
	http.Client
	ctx           context.Context
	baseUrl       string
	retryCount    int
	retryInterval time.Duration

	rw          sync.RWMutex
	headers     map[string]string
	cookies     map[string]string
	middlewares []MiddlewareFunc
}

const (
	defaultUserAgent = "DobyteHttpClient"

	HeaderUserAgent     = "User-Agent"
	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"
	HeaderCookie        = "Cookie"
	HeaderHost          = "Host"

	ContentTypeJson           = "application/json"
	ContentTypeXml            = "application/xml"
	ContentTypeFormData       = "form-data"
	ContentTypeFormUrlEncoded = "application/x-www-form-urlencoded"
)

func NewClient() *Client {
	c := &Client{
		Client: http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		headers:     make(map[string]string),
		cookies:     make(map[string]string),
		middlewares: make([]MiddlewareFunc, 0),
	}

	c.SetHeader(HeaderUserAgent, defaultUserAgent)

	return c
}

// SetHeader Set a common header for the client.
func (c *Client) SetHeader(key, value string) {
	c.rw.Lock()
	defer c.rw.Unlock()

	c.headers[key] = value
}

// SetHeaders Set multiple common headers for the client.
func (c *Client) SetHeaders(headers map[string]string) {
	c.rw.Lock()
	defer c.rw.Unlock()

	for key, value := range headers {
		c.headers[key] = value
	}
}

// GetHeaders Returns all common headers.
func (c *Client) GetHeaders() map[string]string {
	c.rw.RLock()
	defer c.rw.RUnlock()

	headers := make(map[string]string, len(c.headers))
	for key, value := range c.headers {
		headers[key] = value
	}

	return headers
}

// SetCookie Set a common cookie for the client.
func (c *Client) SetCookie(key, value string) {
	c.rw.Lock()
	defer c.rw.Unlock()

	c.cookies[key] = value
}

// SetCookies Set multiple common cookies for the client.
func (c *Client) SetCookies(cookies map[string]string) {
	c.rw.Lock()
	defer c.rw.Unlock()

	for key, value := range cookies {
		c.cookies[key] = value
	}
}

// GetCookies Returns all common cookies.
func (c *Client) GetCookies() map[string]string {
	c.rw.RLock()
	defer c.rw.RUnlock()

	cookies := make(map[string]string, len(c.cookies))
	for key, value := range c.cookies {
		cookies[key] = value
	}

	return cookies
}

// SetUserAgent Set User-Agent for the request.
func (c *Client) SetUserAgent(agent string) {
	c.SetHeader(HeaderUserAgent, agent)
}

// SetContentType Set Content-Type for the request.
func (c *Client) SetContentType(contentType string) {
	c.SetHeader(HeaderContentType, contentType)
}

// SetBrowserMode Enable browser mode for the request.
func (c *Client) SetBrowserMode() {
	c.Jar, _ = cookiejar.New(nil)
}

// SetBaseUrl Set base url for the client.
func (c *Client) SetBaseUrl(baseUrl string) {
	c.baseUrl = baseUrl
}

// GetBaseUrl Returns base url of the client.
func (c *Client) GetBaseUrl() string {
	return c.baseUrl
}

// SetBasicAuth Set HTTP basic authentication information for the client.
func (c *Client) SetBasicAuth(username, password string) {
	c.SetHeader(HeaderAuthorization, "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
}

// SetBearerToken Set HTTP Bearer-Token authentication information for the client.
func (c *Client) SetBearerToken(token string) {
	c.SetHeader(HeaderAuthorization, "Bearer "+token)
}

// SetContext Set context for the client.
func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// SetTimeout sets the request timeout for the client.
func (c *Client) SetTimeout(timeout time.Duration) {
	c.Timeout = timeout
}

// SetRetry sets count and interval of retry for the client.
func (c *Client) SetRetry(retryCount int, retryInterval time.Duration) {
	c.retryCount, c.retryInterval = retryCount, retryInterval
}

func (c *Client) SetKeepAlive(enable bool) {
	//c.Transport.
}

// Use sets middleware for the client.
func (c *Client) Use(middlewares ...MiddlewareFunc) {
	c.rw.Lock()
	defer c.rw.Unlock()

	c.middlewares = append(c.middlewares, middlewares...)
}

func (c *Client) getMiddlewares() []MiddlewareFunc {
	c.rw.RLock()
	defer c.rw.RUnlock()

	middlewares := make([]MiddlewareFunc, len(c.middlewares))
	copy(middlewares, c.middlewares)
	return middlewares
}

// Download a file from the remote address to the local.
func (c *Client) Download(url, dir string, filename ...string) (string, error) {
	return newDownload(c).download(url, dir, filename...)
}

// Upload multi files to remote address.
func (c *Client) Upload(url string, files interface{}, data interface{}, opts ...*UploadOptions) (*Response, error) {
	return newUpload(c).request(url, files, data, opts...)
}

// Request send an http request.
func (c *Client) Request(method, url string, data interface{}, opts ...*RequestOptions) (*Response, error) {
	return newRequest(c).request(method, url, data, opts...)
}

// Get Send a http request use get method.
func (c *Client) Get(url string, data interface{}, opts ...*RequestOptions) (*Response, error) {
	return c.Request(MethodGet, url, data, opts...)
}

// Post Send a http request use post method.
func (c *Client) Post(url string, data interface{}, opts ...*RequestOptions) (*Response, error) {
	return c.Request(MethodPost, url, data, opts...)
}

// Put Send a http request use put method.
func (c *Client) Put(url string, data interface{}, opts ...*RequestOptions) (*Response, error) {
	return c.Request(MethodPut, url, data, opts...)
}

// Patch Send a http request use patch method.
func (c *Client) Patch(url string, data interface{}, opts ...*RequestOptions) (*Response, error) {
	return c.Request(MethodPatch, url, data, opts...)
}

// Delete Send a http request use patch method.
func (c *Client) Delete(url string, data interface{}, opts ...*RequestOptions) (*Response, error) {
	return c.Request(MethodDelete, url, data, opts...)
}

// Head Send a http request use head method.
func (c *Client) Head(url string, data interface{}, opts ...*RequestOptions) (*Response, error) {
	return c.Request(MethodHead, url, data, opts...)
}

// Options Send a request use options method.
func (c *Client) Options(url string, data interface{}, opts ...*RequestOptions) (*Response, error) {
	return c.Request(MethodOptions, url, data, opts...)
}

// Connect Send a request use connect method.
func (c *Client) Connect(url string, data interface{}, opts ...*RequestOptions) (*Response, error) {
	return c.Request(MethodConnect, url, data, opts...)
}

// Trace Send a request use trace method.
func (c *Client) Trace(url string, data interface{}, opts ...*RequestOptions) (*Response, error) {
	return c.Request(MethodTrace, url, data, opts...)
}
