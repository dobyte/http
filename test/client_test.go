/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/8/16 2:54 下午
 * @Desc: TODO
 */

package test_test

import (
	"context"
	"fmt"
	"github.com/dobyte/http"
	"io/ioutil"
	stdhttp "net/http"
	"os"
	"sync"
	"testing"
	"time"
)

const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlc2MiOjE2MjgwNDAzMjYxNTQ2MzIwMDAsImV4cCI6MTYyODIyMDMyNiwiaWF0IjoxNjI4MDQwMzI2LCJpZCI6MX0.KM19c6URIih-5SyycYIjNAdSiPKxMQEz3DoROm0N3nw"

func TestClient_Get(t *testing.T) {
	client := http.NewClient()
	client.SetBaseUrl("http://www.baidu.com")
	client.Use(func(r http.Request) (*http.Response, error) {
		return r.Next()
	})

	data := struct {
		Code string `json:"code"`
	}{
		Code: "",
	}

	resp, err := client.Get("/", data)
	if err != nil {
		t.Fatal(err)
	}

	body, err := resp.ReadBody()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(body))
}

func TestClient_Post(t *testing.T) {
	//client := http.NewClient()
	//client.SetBaseUrl("http://127.0.0.1:8199")
	//client.SetBearerToken(token)
	//client.SetContentType(http.ContentTypeJson)
	//client.Use(func(r *http.Request) (*http.Response, error) {
	//	r.Request.Header.Set("Client-Type", "2")
	//	return r.Next()
	//})
	//
	//type updateRegionArg struct {
	//	Id   int    `json:"id"`
	//	Pid  int    `json:"pid"`
	//	Code string `json:"code"`
	//	Name string `json:"name"`
	//	Sort int    `json:"sort"`
	//}
	//
	//data := updateRegionArg{
	//	Id:   1,
	//	Pid:  0,
	//	Code: "110000",
	//	Name: "北京市",
	//	Sort: 0,
	//}
	//
	//if resp, err := client.Put("/backend/region/update-region", data, nil); err != nil {
	//	t.Error(err)
	//	return
	//} else {
	//	t.Log(resp.Response.Status)
	//	t.Log(resp.Response.Header)
	//	t.Log(resp.ReadBytes())
	//	t.Log(resp.ReadString())
	//	t.Log(resp.GetHeaders())
	//	t.Log(resp.GetCookies())
	//}
}

func TestClient_Download(t *testing.T) {
	url := "https://www.baidu.com/img/PCtm_d9c8750bed0b3c7d089fa7d55720d6cf.png"

	if path, err := http.NewClient().Download(url, "./"); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(path)
	}
}

func TestClient_Upload(t *testing.T) {
	client := http.NewClient()
	client.SetBaseUrl("http://127.0.0.1:8899/")

	type Files struct {
		Image           []string `name:"image"`
		BackgroundImage string   `name:"backgroundImage"`
	}

	type Data struct {
		DisplayName string `name:"displayName"`
		Description string `name:"description"`
	}

	files := &Files{
		Image: []string{"./screenshot-20221107-134213.png", "./screenshot-20221107-134832.png"},
	}
	data := &Data{
		DisplayName: "displayName",
		Description: "description",
	}

	_, err := client.Upload("/upload", files, data)
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkClient_Get(b *testing.B) {
	client := http.NewClient()
	client.SetBaseUrl("https://www.baidu.com/")
	client.Use(func(r http.Request) (*http.Response, error) {
		return r.Next()
	})

	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			_, err := client.Request(http.MethodGet, "", nil, nil)
			if err != nil {
				b.Error(err)
			}

			wg.Done()
		}()
	}

	wg.Wait()

	time.Sleep(3 * time.Second)
}

func TestClient_Request(t *testing.T) {
	type fields struct {
		ctx           context.Context
		baseUrl       string
		retryCount    int
		retryInterval time.Duration
		headers       map[string]string
		cookies       map[string]string
		middlewares   []http.MiddlewareFunc
	}
	type args struct {
		method string
		url    string
		data   interface{}
		opts   []*http.RequestOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "simple get request without query params",
			fields: fields{
				baseUrl: "http://117.50.20.231:8080",
			},
			args: args{
				method: http.MethodGet,
				url:    "/notify/announcement",
			},
			wantErr: false,
		},
		{
			name: "simple get request with string query params",
			fields: fields{
				baseUrl: "http://117.50.20.231:8080",
			},
			args: args{
				method: http.MethodGet,
				url:    "/notify/announcement",
				data:   "is_force=true",
			},
			wantErr: false,
		},
		{
			name: "simple get request with struct query params",
			fields: fields{
				baseUrl: "http://117.50.20.231:8080",
			},
			args: args{
				method: http.MethodGet,
				url:    "/notify/announcement",
				data: struct {
					IsForce bool `json:"is_force"`
				}{
					IsForce: true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := http.NewClient()
			c.SetContext(tt.fields.ctx)
			c.SetBaseUrl(tt.fields.baseUrl)
			c.SetRetry(tt.fields.retryCount, tt.fields.retryInterval)
			c.SetHeaders(tt.fields.headers)
			c.SetCookies(tt.fields.cookies)
			c.Use(tt.fields.middlewares...)

			got, err := c.Request(tt.args.method, tt.args.url, tt.args.data, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Request() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			buf, err := got.ReadBody()
			if err != nil {
				t.Errorf("Request() read body failed error = %v", err)
			} else {
				t.Logf("Request() body = %v", string(buf))
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	dir := "./uploads/"
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		t.Fatalf("make dir failed: %v", err)
	}

	server := stdhttp.Server{Addr: ":8899"}

	stdhttp.HandleFunc("/upload", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		r.ParseMultipartForm(64 << 20)

		for _, fileHeaders := range r.MultipartForm.File {
			for _, fileHeader := range fileHeaders {
				file, err := fileHeader.Open()
				if err != nil {
					continue
				}

				buf, err := ioutil.ReadAll(file)
				if err != nil {
					continue
				}

				_ = ioutil.WriteFile(fmt.Sprintf("%s%s", dir, fileHeader.Filename), buf, os.ModePerm)
			}
		}

		w.Write([]byte("ok"))
	})

	server.ListenAndServe()
}
