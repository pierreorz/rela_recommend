package signature

// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

type postFormSignature struct{}

func (postFormSignature) Name() string {
	return "postForm"
}

func (postFormSignature) Signature(req *http.Request) error {
	result, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	form, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return err
	}
	return newSignature(getSignature(req), sortQuery(form), string(result))
}
