package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func get(url string, sid string, result interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if sid != "" {
		req.AddCookie(&http.Cookie{
			Name:  "sid",
			Value: sid,
		})
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 300 {
		return fmt.Errorf("http request failed: %v: %v", res.Status, string(bs))
	}

	if err := json.Unmarshal(bs, result); err != nil {
		return err
	}

	return nil
}

func post(url string, sid string, body interface{}, result interface{}) (*http.Response, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bs))
	if err != nil {
		return nil, err
	}
	if sid != "" {
		req.AddCookie(&http.Cookie{
			Name:  "sid",
			Value: sid,
		})
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	bs, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("http request failed: %v: %v", res.Status, string(bs))
	}

	if result != nil {
		if err := json.Unmarshal(bs, result); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func delete(url string, sid string) error {
	req, err := http.NewRequest(http.MethodDelete, url, http.NoBody)
	if err != nil {
		return err
	}
	if sid != "" {
		req.AddCookie(&http.Cookie{
			Name:  "sid",
			Value: sid,
		})
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 300 {
		return fmt.Errorf("http request failed: %v: %v", res.Status, string(bs))
	}

	return nil
}
