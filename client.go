package confclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	url        string
	httpClient *http.Client
	scopeVars  map[string]string
}

func InitiateClient(url string) *Client {
	client := &Client{
		url:        url,
		httpClient: &http.Client{},
		scopeVars:  make(map[string]string),
	}

	// Get all environment variables beginning with CFG_
	//  And store them in the scopeVars map
	allEnvVars := os.Environ()
	for _, e := range allEnvVars {
		index := strings.Index(e, "=")
		varname := strings.ToLower(e[:index])
		if strings.HasPrefix(varname, "cfg_") {
			scopevar := strings.TrimPrefix(varname, "cfg_")
			log.WithFields(log.Fields{scopevar: e[index+1:]}).Info("Using scope")
			client.scopeVars[scopevar] = e[index+1:]
		}
	}
	return client
}

func (c *Client) GetList(key string) (ListResponse, error) {
	var resp ListResponse

	body, err := c.GETRequestJSON("/list/" + key)
	if err != nil {
		return resp, err
	}

	if err = json.Unmarshal(body, &resp); err != nil {
		return resp, err
	}
	return resp, err
}

func (c *Client) GetHash(key string) (HashResponse, error) {
	var resp HashResponse
	body, err := c.GETRequestJSON("/hash/" + key)
	if err != nil {
		return resp, err
	}

	if err = json.Unmarshal(body, &resp); err != nil {
		return resp, err
	}

	return resp, err
}

func (c *Client) GetString(key string) (StringResponse, error) {

	var resp StringResponse

	body, err := c.GETRequestJSON("/string/" + key)
	if err != nil {
		return resp, err
	}

	if err = json.Unmarshal(body, &resp); err != nil {
		return resp, err
	}

	return resp, err
}

func (c *Client) GETRequestTEXT(path string) ([]byte, error) {
	return c.GETRequest(path, "text/plain")
}

func (c *Client) GETRequestJSON(path string) ([]byte, error) {
	return c.GETRequest(path, "application/json")
}

func (c *Client) GETRequest(path string, accept string) ([]byte, error) {
	req_url := strings.Join([]string{c.url, path}, "")

	req, err := http.NewRequest("GET", req_url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)

	// Send scope variables as x-cfg-blah request headers
	for scopeKey, scopeVal := range c.scopeVars {
		req.Header.Add("x-cfg-"+scopeKey, scopeVal)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	l := log.WithFields(log.Fields{"url": req_url, "httpcode": resp.StatusCode, "method": "GET"})
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		l.Warn("HTTP error")
		return nil, errors.New(fmt.Sprintf("HTTP Error %d", resp.StatusCode))
	}
	l.Debug("HTTP log")

	return ioutil.ReadAll(resp.Body)

}

func (c *Client) PATCHRequestJSON(path string, data []byte) ([]byte, error) {
	req_url := strings.Join([]string{c.url, path}, "")
	req, err := http.NewRequest("PATCH", req_url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	l := log.WithFields(log.Fields{"url": req_url, "httpcode": resp.StatusCode, "method": "POST"})

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		l.Warn("HTTP error")
		return nil, errors.New(fmt.Sprintf("HTTP Error %d", resp.StatusCode))
	}
	l.Debug("HTTP log")

	return ioutil.ReadAll(resp.Body)
}

func (c *Client) POSTRequestJSON(path string, data []byte) ([]byte, error) {
	req_url := strings.Join([]string{c.url, path}, "")
	req, err := http.NewRequest("POST", req_url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	l := log.WithFields(log.Fields{"url": req_url, "httpcode": resp.StatusCode, "method": "POST"})

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		l.Warn("HTTP error")
		return nil, errors.New(fmt.Sprintf("HTTP Error %d", resp.StatusCode))
	}
	l.Debug("HTTP log")

	return ioutil.ReadAll(resp.Body)
}

func (c *Client) DELETERequest(path string) ([]byte, error) {
	req_url := strings.Join([]string{c.url, path}, "")
	req, err := http.NewRequest("DELETE", req_url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	l := log.WithFields(log.Fields{"url": req_url, "httpcode": resp.StatusCode, "method": "DELETE"})

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		l.Warn("HTTP error")
		return nil, errors.New(fmt.Sprintf("HTTP Error %d", resp.StatusCode))
	}
	l.Debug("HTTP log")
	return ioutil.ReadAll(resp.Body)
}
