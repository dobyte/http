package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/dobyte/http/internal/multipart"
	"github.com/dobyte/http/internal/xfile"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	FieldTypeNone           = multipart.FieldTypeNone
	FieldTypeJson           = multipart.FieldTypeJson
	FieldTypeXml            = multipart.FieldTypeXml
	FieldTypeFormUrlEncoded = multipart.FieldTypeFormUrlEncoded
)

type FieldType = multipart.FieldType

type UploadOptions struct {
	Headers   map[string]string
	Cookies   map[string]string
	FieldType FieldType
}

type upload struct {
	executor
}

func newUpload(client *Client) *upload {
	return &upload{executor{client: client}}
}

// send a http request.
func (r *upload) request(url string, files, data interface{}, opts ...*UploadOptions) (*Response, error) {
	req, err := r.prepare(url, files, data, opts...)
	if err != nil {
		return nil, err
	}

	return r.call(req)
}

// build a http request.
func (r *upload) prepare(url string, files, data interface{}, opts ...*UploadOptions) (req *http.Request, err error) {
	var (
		buffer  = &bytes.Buffer{}
		writer  = multipart.NewWriter(buffer)
		headers = r.client.GetHeaders()
		cookies = r.client.GetCookies()
	)

	if err = r.writeFiles(writer, files); err != nil {
		return
	}

	fieldType := FieldTypeNone
	if len(opts) > 0 {
		fieldType = opts[0].FieldType
	}

	if err = r.writeData(writer, data, fieldType); err != nil {
		return
	}

	_ = writer.Close()

	req, err = http.NewRequest(MethodPost, r.makeUrl(url), buffer)
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
		case HeaderContentType, HeaderCookie:
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

	req.Host = req.Header.Get(HeaderHost)
	req.Header.Set(HeaderContentType, writer.FormDataContentType())

	return
}

func (r *upload) writeFiles(writer *multipart.Writer, files interface{}) error {
	set := make(fileset)
	switch v := files.(type) {
	case map[string]string:
		for name, path := range v {
			set.add(name, path)
		}
	case *map[string]string:
		for name, path := range *v {
			set.add(name, path)
		}
	case map[string][]string:
		for name, paths := range v {
			for _, path := range paths {
				set.add(name, path)
			}
		}
	case *map[string][]string:
		for name, paths := range *v {
			for _, path := range paths {
				set.add(name, path)
			}
		}
	default:
		var (
			rv   = reflect.ValueOf(files)
			kind = rv.Kind()
		)

		for kind == reflect.Ptr {
			rv = rv.Elem()
			kind = rv.Kind()
		}

		switch kind {
		case reflect.Map:
			iter := rv.MapRange()
			for iter.Next() {
				switch iv := iter.Value(); iv.Kind() {
				case reflect.String:
					set.add(iter.Key().String(), iv.String())
				case reflect.Slice, reflect.Array:
					name := iter.Key().String()
					for n := 0; n < iv.Len(); n++ {
						switch itv := iv.Index(n); itv.Kind() {
						case reflect.String:
							set.add(name, itv.String())
						}
					}
				}
			}
		case reflect.Struct:
			var (
				name string
				rt   = rv.Type()
			)

			for i := 0; i < rv.NumField(); i++ {
				name = rt.Field(i).Tag.Get("name")
				if name == "" {
					name = rt.Field(i).Name
				}

				switch iv := rv.Field(i); iv.Kind() {
				case reflect.String:
					set.add(name, iv.String())
				case reflect.Slice, reflect.Array:
					for n := 0; n < iv.Len(); n++ {
						switch itv := iv.Index(n); itv.Kind() {
						case reflect.String:
							set.add(name, itv.String())
						}
					}
				}
			}
		default:
			return errors.New("files type must be map or struct")
		}
	}

	var (
		err    error
		file   *os.File
		stream io.Writer
	)

	for name, paths := range set {
		for _, path := range paths {
			if !xfile.Exists(path) {
				return errors.New(fmt.Sprintf(`"%s" does not exist`, path))
			}

			stream, err = writer.CreateFormFile(name, filepath.Base(path))
			if err != nil {
				return err
			}

			file, err = os.Open(path)
			if err != nil {
				return err
			}

			_, err = io.Copy(stream, file)
			_ = file.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *upload) writeData(writer *multipart.Writer, data interface{}, fieldType FieldType) (err error) {
	if data == nil {
		return
	}

	var (
		rv   = reflect.ValueOf(data)
		kind = rv.Kind()
	)

	for kind == reflect.Ptr {
		rv = rv.Elem()
		kind = rv.Kind()
	}

	switch kind {
	case reflect.Map:
		iter := rv.MapRange()
		for iter.Next() {
			if err = writer.WriteField(iter.Key().String(), iter.Value().Interface(), fieldType); err != nil {
				return
			}
		}
	case reflect.Struct:
		var (
			rt   = rv.Type()
			ok   bool
			name string
		)
		for i := 0; i < rv.NumField(); i++ {
			for _, tag := range []string{"name", "field", "json", "xml"} {
				name, ok = rt.Field(i).Tag.Lookup(tag)
				if ok {
					break
				}
			}
			if name == "" {
				name = rt.Field(i).Name
			}

			if err = writer.WriteField(name, rv.Field(i).Interface(), fieldType); err != nil {
				return
			}
		}
	default:
		err = errors.New("data type must be map or struct")
	}

	return
}

type fileset map[string][]string

func (fs fileset) add(name string, path string) {
	if path == "" {
		return
	}

	if _, ok := fs[name]; !ok {
		fs[name] = make([]string, 0, 1)
		fs[name] = append(fs[name], path)
	} else {
		fs[name] = append(fs[name], path)
	}
}
