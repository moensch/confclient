package confclient

import (
	"encoding/json"
	"errors"
	"fmt"
)

type KeyResponse struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
type StringKeyResponse struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type SimpleListResponse struct {
	Data []string `json:"data"`
}

type DataRequest struct {
	Data string `json:"data"`
}

func (c *Client) AdminGetKeyAsTEXT(keyName string) (string, error) {
	resp, err := c.GETRequestTEXT(fmt.Sprintf("/admin/key/%s", keyName))
	if err != nil {
		return "", err
	}

	return string(resp), err
}

func (c *Client) AdminGetKeyAsJSON(keyName string) ([]byte, error) {
	var jsonblob = make([]byte, 0)

	resp, err := c.GETRequestJSON(fmt.Sprintf("/admin/key/%s", keyName))
	if err != nil {
		return jsonblob, err
	}
	var keyResponse KeyResponse
	if err := json.Unmarshal(resp, &keyResponse); err != nil {
		return jsonblob, err
	}
	jsonblob, err = json.MarshalIndent(keyResponse, "", "  ")
	if err != nil {
		return jsonblob, err
	}

	return jsonblob, err
}

func (c *Client) AdminSetKeyFromJSON(keyName string, jsonblob []byte) error {
	_, err := c.POSTRequestJSON(fmt.Sprintf("/admin/key/%s", keyName), jsonblob)
	return err
}

func (c *Client) AdminGetHashField(keyName string, fieldName string) (string, error) {
	resp, err := c.GETRequestJSON(fmt.Sprintf("/admin/key/%s/%s", keyName, fieldName))
	if err != nil {
		return "", err
	}

	var keyResponse KeyResponse
	if err := json.Unmarshal(resp, &keyResponse); err != nil {
		return "", err
	}
	return keyResponse.Data.(string), err
}

func (c *Client) AdminGetListIndex(keyName string, index string) (string, error) {
	resp, err := c.GETRequestJSON(fmt.Sprintf("/admin/key/%s/index/%s", keyName, index))
	if err != nil {
		return "", err
	}

	var keyResponse KeyResponse
	if err := json.Unmarshal(resp, &keyResponse); err != nil {
		return "", err
	}
	return keyResponse.Data.(string), err
}

func (c *Client) AdminSetStringKey(keyName string, value string) error {
	ktype, err := c.AdminGetKeyType(keyName)
	if err != nil {
		return err
	}
	if ktype != "string" {
		return errors.New(fmt.Sprintf("Can only set keys of type 'string' via parameter - type '%s' not supported!", ktype))
	}
	req := &StringKeyResponse{
		Type: "string",
		Data: value,
	}
	jsonblob, err := json.Marshal(req)
	err = c.AdminSetKeyFromJSON(keyName, jsonblob)

	return err
}

func (c *Client) AdminSetHashField(keyName string, fieldName string, value string) error {
	req := &DataRequest{value}

	jsonblob, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = c.POSTRequestJSON(fmt.Sprintf("/admin/key/%s/%s", keyName, fieldName), jsonblob)
	return err
}

func (c *Client) AdminDeleteKey(keyName string) error {
	_, err := c.DELETERequest(fmt.Sprintf("/admin/key/%s", keyName))
	return err
}

func (c *Client) AdminListKeys(filter string) ([]string, error) {
	keys := make([]string, 0)

	var req_url string
	if filter != "" {
		req_url = fmt.Sprintf("/admin/keys/%s", filter)
	} else {
		req_url = "/admin/keys"
	}

	var resp SimpleListResponse

	body, err := c.GETRequestJSON(req_url)
	if err != nil {
		return keys, err
	}

	if err = json.Unmarshal(body, &resp); err != nil {
		return keys, err
	}
	return resp.Data, err
}

func (c *Client) AdminListHashFields(keyName string) ([]string, error) {
	fields := make([]string, 0)

	var resp SimpleListResponse

	body, err := c.GETRequestJSON(fmt.Sprintf("/admin/util/hashfields/%s", keyName))
	if err != nil {
		return fields, err
	}

	if err = json.Unmarshal(body, &resp); err != nil {
		return fields, err
	}
	return resp.Data, err
}

func (c *Client) AdminGetKeyType(keyName string) (string, error) {
	resp, err := c.GETRequestJSON(fmt.Sprintf("/admin/util/type/%s", keyName))
	if err != nil {
		return "", err
	}

	var keyResponse KeyResponse
	if err := json.Unmarshal(resp, &keyResponse); err != nil {
		return "", err
	}
	return keyResponse.Data.(string), err
}
