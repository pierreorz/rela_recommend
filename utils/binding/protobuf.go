// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"github.com/golang/protobuf/proto"

	"io/ioutil"
	"net/http"
	"net/url"
)

type protobufBinding struct{}

func (protobufBinding) Name() string {
	return "protobuf"
}

func (protobufBinding) Bind(req *http.Request, obj interface{}) error {
	form, _ := url.ParseQuery(req.URL.RawQuery)
	if err := mapForm(obj, form); err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	if err = proto.Unmarshal(buf, obj.(proto.Message)); err != nil {
		return err
	}

	//Here it's same to return validate(obj), but util now we cann't add `binding:""` to the struct
	//which automatically generate by gen-proto
	return nil
	//return validate(obj)
}
