package signature

import (
	"errors"
	"net/http"
	"sort"
	"strings"

	//"theL_api_golang/log"
	"rela_recommend/utils"
)

const (
	SALT      = "MAaFS5Zc6ZIEapnmhurNyLLFwf3xWm"
	NewSalt   = "879f30c4b1641142c6192acc23cfb733"
	SIGNATURE = "signature"
)

const (
	MIMEJSON = "application/json"

	HeaderContentType = "Content-Type"
	HeaderSignature   = "Signature"
)

type ISignature interface {
	Name() string
	Signature(*http.Request) error
}

var (
	GetForm          = getFormSignature{}
	PostForm         = postFormSignature{}
	Json             = jsonSignature{}
	ErrInvaSignature = errors.New("invalid signature")
)

func Default(method, contentType string) ISignature {
	if method == http.MethodGet {
		return GetForm
	}
	switch contentType {
	case MIMEJSON:
		return Json
	default:
		return PostForm
	}
}

func Signature(req *http.Request) error {
	if getSignature(req) != "" { //新版签名
		return Default(req.Method, getContentType(req)).Signature(req)
	}

	if err := req.ParseForm(); err != nil {
		return err
	}
	return signature(req.Form)
}

func signature(form map[string][]string) error {
	inputSignature, exists := form[SIGNATURE]
	if !exists {
		return ErrInvaSignature
	}

	values := make([]string, 0)
	for key, value := range form {
		if key != SIGNATURE {
			numElems := len(value)
			for i := 0; i < numElems; i++ {
				values = append(values, key+"="+value[i])
			}
		}
	}
	sort.Strings(values)

	//log.Info(strings.Join(values, "&") + SALT)

	newSignature := utils.Md5Sum([]byte(strings.Join(values, "&") + SALT))
	if newSignature != inputSignature[0] {
		return ErrInvaSignature
	}
	return nil
}

func sortQuery(form map[string][]string) string {
	values := make([]string, 0)
	for key, value := range form {
		numElems := len(value)
		for i := 0; i < numElems; i++ {
			values = append(values, key+"="+value[i])
		}
	}
	sort.Strings(values)
	return strings.Join(values, "&")
}

func newSignature(signature, sortQueryString, body string) error {
	newSignature := utils.Md5Sum([]byte(sortQueryString + body + NewSalt))
	if newSignature != signature {
		return ErrInvaSignature
	}
	return nil
}

func getSignature(req *http.Request) string {
	return req.Header.Get(HeaderSignature)
}

func getContentType(req *http.Request) string {
	return req.Header.Get(HeaderContentType)
}
