// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"encoding/xml"
	"net/http"
	"net/url"
)

type xmlBinding struct{}

func (xmlBinding) Name() string {
	return "xml"
}

func (xmlBinding) Bind(req *http.Request, obj interface{}) error {
	form, _ := url.ParseQuery(req.URL.RawQuery)
	if err := mapForm(obj, form); err != nil {
		return err
	}

	decoder := xml.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validate(obj)
}
