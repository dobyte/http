package multipart

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/dobyte/http/internal"
	"io"
	"mime/multipart"
	"net/textproto"
	"reflect"
	"strconv"
	"strings"
)

type FieldType string

const (
	FieldTypeNone           FieldType = ""
	FieldTypeJson           FieldType = "application/json"
	FieldTypeXml            FieldType = "application/xml"
	FieldTypeFormUrlEncoded FieldType = "application/x-www-form-urlencoded"
)

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

type Writer struct {
	*multipart.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{Writer: multipart.NewWriter(w)}
}

func (w *Writer) WriteField(fieldName string, fieldValue interface{}, fieldType FieldType) (err error) {
	var (
		buf  []byte
		rv   = reflect.ValueOf(fieldValue)
		kind = rv.Kind()
	)

	for kind == reflect.Ptr {
		rv = rv.Elem()
		kind = rv.Kind()
	}

	switch kind {
	case reflect.Bool:
		buf = []byte(strconv.FormatBool(rv.Bool()))
		fieldType = FieldTypeNone
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf = []byte(strconv.FormatInt(rv.Int(), 10))
		fieldType = FieldTypeNone
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		buf = []byte(strconv.FormatUint(rv.Uint(), 10))
		fieldType = FieldTypeNone
	case reflect.Float32, reflect.Float64:
		buf = []byte(strconv.FormatFloat(rv.Float(), 'f', -1, 64))
		fieldType = FieldTypeNone
	case reflect.Complex64, reflect.Complex128:
		buf = []byte(strconv.FormatComplex(rv.Complex(), 'e', -1, 64))
		fieldType = FieldTypeNone
	case reflect.String:
		buf = []byte(rv.String())
		fieldType = FieldTypeNone
	case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct, reflect.Interface:
		switch fieldType {
		case FieldTypeXml:
			buf, err = xml.Marshal(rv.Interface())
			if err != nil {
				return
			}
		case FieldTypeFormUrlEncoded:
			buf = []byte(internal.BuildParams(rv.Interface()))
		default:
			buf, err = json.Marshal(rv.Interface())
			if err != nil {
				return
			}
			fieldType = FieldTypeJson
		}
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, quoteEscaper.Replace(fieldName)))

	if fieldType != FieldTypeNone {
		h.Set("Content-Type", string(fieldType))
	}

	p, err := w.CreatePart(h)
	if err != nil {
		return
	}

	_, err = p.Write(buf)
	return err
}
