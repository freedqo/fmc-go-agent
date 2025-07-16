package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"github.com/gocarina/gocsv"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"reflect"
	"strings"
)

func CreatCsvFilebytesBuffer(in interface{}) (*bytes.Buffer, error) {
	//组装映射设备点表csv
	devbytes, err := gocsv.MarshalBytes(in)
	if err != nil {
		return nil, err
	}
	gbkDevReader := transform.NewReader(bytes.NewReader(devbytes), simplifiedchinese.GBK.NewEncoder())
	gbkbuf1 := &bytes.Buffer{}
	_, err = gbkbuf1.ReadFrom(gbkDevReader)
	if err != nil {
		return nil, err
	}
	return gbkbuf1, err
}

func FileBytesToGzipBase64Str(inbytes *bytes.Buffer) (string, error) {
	gzipbuf := &bytes.Buffer{}
	w := gzip.NewWriter(gzipbuf)
	_, err := w.Write(inbytes.Bytes())
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}
	//转Base64字符串
	base64str1 := base64.StdEncoding.EncodeToString(gzipbuf.Bytes())
	return base64str1, nil
}

func BoolToYn(flag bool) string {
	if flag {
		return "Y"
	} else {
		return "N"
	}
}

func YnToBool(instr string) bool {
	if strings.ToUpper(strings.TrimSpace(instr)) == "Y" {
		return true
	} else {
		return false
	}
}

func GetStructTagString(in interface{}, tagName string) ([]string, error) {

	tagStringlist := make([]string, 0)
	typ := reflect.TypeOf(in)
	val := reflect.ValueOf(in)

	kd := val.Kind() //获取到a对应的类别
	if kd != reflect.Struct {
		return tagStringlist, errors.New("入参interface不是结构体")
	}
	//获取到该结构体有几个字段
	num := val.NumField()
	//遍历结构体的所有字段
	for i := 0; i < num; i++ {
		//获取到struct标签，需要通过reflect.Type来获取tag标签的值
		tagVal := typ.Field(i).Tag.Get(tagName)
		tagStringlist = append(tagStringlist, tagVal)
	}
	return tagStringlist, nil
}

