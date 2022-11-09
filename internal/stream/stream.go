/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/8/26 6:21 下午
 * @Desc: TODO
 */

package stream

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"strings"
	"sync"
)

var fileTypeMap sync.Map

func init() {
	fileTypeMap.Store("ffd8ffe000104a464946", "jpg")
	fileTypeMap.Store("89504e470d0a1a0a0000", "png")
	fileTypeMap.Store("47494638396126026f01", "gif")
	fileTypeMap.Store("49492a00227105008037", "tif")
	fileTypeMap.Store("424d228c010000000000", "bmp")
	fileTypeMap.Store("424d8240090000000000", "bmp")
	fileTypeMap.Store("424d8e1b030000000000", "bmp")
	fileTypeMap.Store("41433130313500000000", "dwg")
	fileTypeMap.Store("3c21444f435459504520", "html")
	fileTypeMap.Store("3c68746d6c3e0", "html")
	fileTypeMap.Store("3c21646f637479706520", "htm")
	fileTypeMap.Store("48544d4c207b0d0a0942", "css")
	fileTypeMap.Store("696b2e71623d696b2e71", "js")
	fileTypeMap.Store("7b5c727466315c616e73", "rtf")
	fileTypeMap.Store("38425053000100000000", "psd")
	fileTypeMap.Store("46726f6d3a203d3f6762", "eml")
	fileTypeMap.Store("d0cf11e0a1b11ae10000", "doc")
	fileTypeMap.Store("d0cf11e0a1b11ae10000", "vsd")
	fileTypeMap.Store("5374616E64617264204A", "mdb")
	fileTypeMap.Store("252150532D41646F6265", "ps")
	fileTypeMap.Store("255044462d312e350d0a", "pdf")
	fileTypeMap.Store("2e524d46000000120001", "rmvb")
	fileTypeMap.Store("464c5601050000000900", "flv")
	fileTypeMap.Store("00000020667479706d70", "mp4")
	fileTypeMap.Store("49443303000000002176", "mp3")
	fileTypeMap.Store("000001ba210001000180", "mpg")
	fileTypeMap.Store("3026b2758e66cf11a6d9", "wmv")
	fileTypeMap.Store("52494646e27807005741", "wav")
	fileTypeMap.Store("52494646d07d60074156", "avi")
	fileTypeMap.Store("4d546864000000060001", "mid")
	fileTypeMap.Store("504b0304140000000800", "zip")
	fileTypeMap.Store("526172211a0700cf9073", "rar")
	fileTypeMap.Store("235468697320636f6e66", "ini")
	fileTypeMap.Store("504b03040a0000000000", "jar")
	fileTypeMap.Store("4d5a9000030000000400", "exe")
	fileTypeMap.Store("3c25402070616765206c", "jsp")
	fileTypeMap.Store("4d616e69666573742d56", "mf")
	fileTypeMap.Store("3c3f786d6c2076657273", "xml")
	fileTypeMap.Store("494e5345525420494e54", "sql")
	fileTypeMap.Store("7061636b616765207765", "java")
	fileTypeMap.Store("406563686f206f66660d", "bat")
	fileTypeMap.Store("1f8b0800000000000000", "gz")
	fileTypeMap.Store("6c6f67346a2e726f6f74", "properties")
	fileTypeMap.Store("cafebabe0000002e0041", "class")
	fileTypeMap.Store("49545346030000006000", "chm")
	fileTypeMap.Store("04000000010000001300", "mxp")
	fileTypeMap.Store("504b0304140006000800", "docx")
	fileTypeMap.Store("d0cf11e0a1b11ae10000", "wps")
	fileTypeMap.Store("6431303a637265617465", "torrent")
	fileTypeMap.Store("6D6F6F76", "mov")
	fileTypeMap.Store("FF575043", "wpd")
	fileTypeMap.Store("CFAD12FEC5FD746F", "dbx")
	fileTypeMap.Store("2142444E", "pst")
	fileTypeMap.Store("AC9EBD8F", "qdf")
	fileTypeMap.Store("E3828596", "pwl")
	fileTypeMap.Store("2E7261FD", "ram")
}

// bytesToHexString get the binary of the previous result byte.
func bytesToHexString(stream []byte) string {
	if stream == nil || len(stream) <= 0 {
		return ""
	}

	var (
		hv   string
		res  = bytes.Buffer{}
		temp = make([]byte, 0)
	)

	for _, v := range stream {
		if hv = hex.EncodeToString(append(temp, v&0xFF)); len(hv) < 2 {
			res.WriteString(strconv.FormatInt(int64(0), 10))
		}

		res.WriteString(hv)
	}

	return res.String()
}

// GetFileType judge the file type based on the binary byte stream.
func GetFileType(stream []byte) string {
	var (
		fileType string
		fileCode = bytesToHexString(stream)
	)

	fileTypeMap.Range(func(key, value interface{}) bool {
		if strings.HasPrefix(fileCode, strings.ToLower(key.(string))) ||
			strings.HasPrefix(key.(string), strings.ToLower(fileCode)) {
			fileType = value.(string)
			return false
		}

		return true
	})

	return fileType
}
