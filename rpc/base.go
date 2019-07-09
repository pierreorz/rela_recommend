package rpc


import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
	"rela_recommend/log"
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
	finalUrl := cli.api_host+url
	req, _ := http.NewRequest(http.MethodPost, finalUrl, bytes.NewBuffer(body))
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
	if err == nil {
		err = json.Unmarshal(data, internalClientRes)
	}

	if err != nil {
		log.Errorf("SendPOSTJson err, %s %s\n", finalUrl, err)
	}
	return err
}

func (cli *HttpClient) SendGETForm(url, params string, internalClientRes interface{}) error {
	finalUrl := cli.api_host+url+"?"+params
	req, _ := http.NewRequest(http.MethodGet, finalUrl, nil)
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
	if err == nil {
		err = json.Unmarshal(data, internalClientRes)
	}

	if err != nil {
		log.Errorf("SendGETForm err, %s %s\n", finalUrl, err)
	}
	return err
}

func (cli *HttpClient) SendPOSTForm(url string, body []byte, internalClientRes interface{}) error {
	finalUrl := cli.api_host+url
	req, _ := http.NewRequest(http.MethodPost, finalUrl, bytes.NewBuffer(body))
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
	if err == nil {
		err = json.Unmarshal(data, internalClientRes)
	}

	if err != nil {
		log.Errorf("SendPOSTForm err, %s %s\n", finalUrl, err)
	}
	return err
}

func (cli *HttpClient) SendGetHeader(url, headerKey, headerVal string, internalClientRes interface{}) error {
	finalUrl := cli.api_host+url
	req, _ := http.NewRequest(http.MethodGet, finalUrl, nil)
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
	if err == nil {
		err = json.Unmarshal(data, internalClientRes)
	}
	if err != nil {
		log.Errorf("SendGetHeader err, %s %s\n", finalUrl, err)
	}
	return err
}
