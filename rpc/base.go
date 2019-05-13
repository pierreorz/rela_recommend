package rpc


import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
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

func (cli *HttpClient) SendPOSTJson(url string, body []byte, internalClientRes interface{}) error {
	req, _ := http.NewRequest(http.MethodPost, cli.api_host+url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := cli.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	if internalClientRes == nil {
		return nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, internalClientRes)
}

func (cli *HttpClient) SendGETForm(url, params string, internalClientRes interface{}) error {
	req, _ := http.NewRequest(http.MethodGet, cli.api_host+url+"?"+params, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := cli.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	if internalClientRes == nil {
		return nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, internalClientRes)
}

func (cli *HttpClient) SendPOSTForm(url string, body []byte, internalClientRes interface{}) error {
	req, _ := http.NewRequest(http.MethodPost, cli.api_host+url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := cli.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	if internalClientRes == nil {
		return nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, internalClientRes)
}

func (cli *HttpClient) SendGetHeader(url, headerKey, headerVal string, internalClientRes interface{}) error {
	req, _ := http.NewRequest(http.MethodGet, cli.api_host+url, nil)
	req.Header.Set(headerKey, headerVal)
	resp, err := cli.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	if internalClientRes == nil {
		return nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, internalClientRes)
}
