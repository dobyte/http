/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/8/26 4:59 下午
 * @Desc: TODO
 */

package xfile

import (
	"io"
	"os"
	"path/filepath"
)

// Exists check if the file or path exists.
func Exists(path string) bool {
	if stat, err := os.Stat(path); stat != nil && !os.IsNotExist(err) {
		return true
	}
	return false
}

// Base returns the last element of path.
// Trailing path separators are removed before extracting the last element.
// If the path is empty, Base returns ".".
// If the path consists entirely of separators, Base returns a single separator.
func Base(path string) string {
	return filepath.Base(path)
}

// SaveToFile save data to file.
func SaveToFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if !Exists(dir) {
		if err := MakeDir(dir); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0666))
	if err != nil {
		return err
	}
	defer f.Close()

	if n, err := f.Write(data); err != nil {
		return err
	} else if n < len(data) {
		return io.ErrShortWrite
	}

	return nil
}

// MakeDir create directories recursively.
func MakeDir(dir string) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// RealPath get the real path.
func RealPath(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	if !Exists(p) {
		return ""
	}
	return p
}
