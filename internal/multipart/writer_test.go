package multipart_test

import (
	"bytes"
	"github.com/dobyte/http/internal/multipart"
	"testing"
)

type User struct {
	ID   int    `json:"id" xml:"id"`
	Name string `json:"name" xml:"name"`
}

func TestWriter_WriteField(t *testing.T) {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)
	//user := &User{
	//	ID:   1,
	//	Name: "fuxiao",
	//}
	users := []*User{{
		ID:   1,
		Name: "fuxiao",
	}}

	err := writer.WriteField("users", users, multipart.FieldTypeXml)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(buffer.String())
}
