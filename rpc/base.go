package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"rela_recommend/log"
	"time"
)

type HttpClient struct {
	api_host string
	client   http.Client
}

func NewHttpClient(api_host string, timeout time.Duration) *HttpClient {
	return &HttpClient{
		api_host: api_host,
		client: http.Client{
			Timeout: timeout,
		},
	}
}

func (cli *HttpClient) doRequest(req *http.Request, internalClientRes interface{}) error {
	resp, err := cli.client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			err = errors.New(resp.Status)
		}
	}
	if internalClientRes == nil {
		return nil
	}
	var data = make([]byte, 0)
	if err == nil {
		data, err = ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(data, internalClientRes)
		}
	}

	if err != nil {
		log.Errorf("doRequest err, %s %s %s %s\n", req.Method, req.URL.String(), string(data), err)
	}
	return err
}

func (cli *HttpClient) SendPOSTJson(url string, body []byte, internalClientRes interface{}) error {
	finalUrl := cli.api_host + url
	req, _ := http.NewRequest(http.MethodPost, finalUrl, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	return cli.doRequest(req, internalClientRes)
}

func (cli *HttpClient) PaiSendPOSTJson(url string, body []byte, internalClientRes interface{}) error {
	finalUrl := cli.api_host + url
	req, _ := http.NewRequest(http.MethodPost, finalUrl, bytes.NewBuffer(body))
	req.Header.Set("Authorization","YjJiYzkyOGM3ZmI3YWI1MjBhZTRhOWY2OTRjZTU4MjMwZTVkNDc2Mg==")
	req.Header.Set("Content-Type", "application/json")
	return cli.doRequest(req, internalClientRes)
}

func (cli *HttpClient) PaiPreSendPOSTJson(url string, body []byte, internalClientRes interface{}) error {
	finalUrl := cli.api_host + url
	req, _ := http.NewRequest(http.MethodPost, finalUrl, bytes.NewBuffer(body))
	req.Header.Set("Authorization","ZTdlM2IwM2Y3ZDllODVkNDViNDFlOGY4YTE5YmI2NjIzMjdmNGMzNw==")
	req.Header.Set("Content-Type", "application/json")
	return cli.doRequest(req, internalClientRes)
}


func (cli *HttpClient) SendGETForm(url, params string, internalClientRes interface{}) error {
	finalUrl := cli.api_host + url + "?" + params
	req, _ := http.NewRequest(http.MethodGet, finalUrl, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return cli.doRequest(req, internalClientRes)
}

func (cli *HttpClient) SendPOSTForm(url string, body []byte, internalClientRes interface{}) error {
	finalUrl := cli.api_host + url
	req, _ := http.NewRequest(http.MethodPost, finalUrl, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return cli.doRequest(req, internalClientRes)
}

func (cli *HttpClient) SendGetHeader(url, headerKey, headerVal string, internalClientRes interface{}) error {
	finalUrl := cli.api_host + url
	req, _ := http.NewRequest(http.MethodGet, finalUrl, nil)
	req.Header.Set(headerKey, headerVal)
	return cli.doRequest(req, internalClientRes)
}
