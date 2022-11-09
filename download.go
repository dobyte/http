/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/8/26 1:59 下午
 * @Desc: TODO
 */

package http

import (
	"github.com/dobyte/http/internal/rand"
	"github.com/dobyte/http/internal/stream"
	"github.com/dobyte/http/internal/xfile"
	"os"
	"strings"
)

var contentTypeToFileSuffix = map[string]string{
	"application/x-001":              ".001",
	"text/h323":                      ".323",
	"drawing/907":                    ".907",
	"audio/x-mei-aac":                ".acp",
	"audio/aiff":                     ".aif",
	"text/asa":                       ".asa",
	"text/asp":                       ".asp",
	"audio/basic":                    ".au",
	"application/vnd.adobe.workflow": ".awf",
	"application/x-bmp":              ".bmp",
	"application/x-c4t":              ".c4t",
	"application/x-cals":             ".cal",
	"application/x-netcdf":           ".cdf",
	"application/x-cel":              ".cel",
	"application/x-g4":               ".cg4",
	"application/x-cit":              ".cit",
	"text/xml":                       ".cml",
	"application/x-cmx":              ".cmx",
	"application/pkix-crl":           ".crl",
	"application/x-csi":              ".csi",
	"application/x-cut":              ".cut",
	"application/x-dbm":              ".dbm",
}

type download struct {
	request *request
}

func newDownload(c *Client) *download {
	return &download{request: newRequest(c)}
}

// Download a file from the network address to the local.
func (d *download) download(url, dir string, filename ...string) (string, error) {
	resp, err := d.request.request(MethodGet, url, nil, nil)
	if err != nil {
		return "", err
	}

	buf, err := resp.ReadBody()
	if err != nil {
		return "", nil
	}

	var path string
	if len(filename) > 0 {
		path = strings.TrimRight(dir, string(os.PathSeparator)) + string(os.PathSeparator) + filename[0]
	} else {
		path = d.genFilePath(buf, dir)
	}

	if err = xfile.SaveToFile(path, buf); err != nil {
		return "", err
	}

	return path, nil
}

// genFilePath generate file path based on response content type
func (d *download) genFilePath(buf []byte, dir string) string {
	path := strings.TrimRight(dir, string(os.PathSeparator)) + string(os.PathSeparator) + rand.Str(16)

	if suffix := stream.GetFileType(buf); suffix != "" {
		path += "." + suffix
	}

	if xfile.Exists(path) {
		return d.genFilePath(buf, dir)
	}

	return path
}
