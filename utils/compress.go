package utils

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

// 压缩类
type ICompress interface {
	Compress(byts []byte) ([]byte, error)
	Decompress(byts []byte) ([]byte, error)
}

type Gzip struct{}

func (self *Gzip) Compress(byts []byte) ([]byte, error) {
	if byts == nil || len(byts) == 0 {
		return byts, nil
	}
	var b bytes.Buffer
	var err error
	gz := gzip.NewWriter(&b)
	defer gz.Close()
	if _, err = gz.Write(byts); err == nil {
		if err = gz.Close(); err == nil {
			if err = gz.Flush(); err == nil {
				return b.Bytes(), nil
			}
		}
	}
	return b.Bytes(), err
}

func (self *Gzip) Decompress(byts []byte) ([]byte, error) {
	if byts == nil || len(byts) == 0 {
		return byts, nil
	}
	var res = []byte{}
	var b = bytes.NewReader(byts)
	gz, err := gzip.NewReader(b)
	defer gz.Close()
	if err == nil {
		if err = gz.Close(); err == nil {
			if res, err = ioutil.ReadAll(gz); err == nil {
				return res, nil
			}
		}
	}
	return res, err
}
