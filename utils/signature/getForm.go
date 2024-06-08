package signature

// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import (
	"net/http"
	"net/url"
)

type getFormSignature struct{}

func (getFormSignature) Name() string {
	return "getForm"
}

func (getFormSignature) Signature(req *http.Request) error {
	form, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return err
	}
	return newSignature(getSignature(req), sortQuery(form), "")

}
