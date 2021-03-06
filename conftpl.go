package confclient

import (
	log "github.com/Sirupsen/logrus"
	"strings"
)

type KeyPair struct {
	Key    string
	Value  string
	Source string
}

type ListResponse struct {
	Type string        `json:"type"`
	Data []ValueSource `json:"data"`
}

type HashResponse struct {
	Type string                 `json:"type"`
	Data map[string]ValueSource `json:"data"`
}

type StringResponse struct {
	Type string      `json:"type"`
	Data ValueSource `json:"data"`
}

type ValueSource struct {
	Value  string `json:"value"`
	Source string `json:"source"`
}

func (c *Client) GetListValueJoined(key string, join_char string) (string, error) {
	list, err := c.GetListValue(key)
	if err != nil {
		return "", err
	}

	return strings.Join(list, join_char), nil
}

func (c *Client) ListExists(key string) (bool, error) {
	_, err := c.GetList(key)

	if err != nil {
		// Does not exist
		return false, nil
	}

	return true, nil
}

func (c *Client) GetListValue(key string) ([]string, error) {
	var strings = make([]string, 0)

	resp, err := c.GetList(key)

	// Errors on list lookups *have* to bubble up
	if err != nil {
		return strings, err
	}

	for idx, v := range resp.Data {
		log.WithFields(log.Fields{
			"key":    key,
			"index":  idx,
			"source": v.Source,
		}).Debug("Got list entry")
		strings = append(strings, v.Value)
	}

	return strings, err
}

func (c *Client) GetListValueDebug(key string) ([]ValueSource, error) {
	var strings = make([]ValueSource, 0)

	resp, err := c.GetList(key)

	// Errors on list lookups *have* to bubble up
	if err != nil {
		return strings, err
	}

	for idx, v := range resp.Data {
		log.WithFields(log.Fields{
			"key":    key,
			"index":  idx,
			"source": v.Source,
		}).Debug("Got list entry")
	}

	return resp.Data, err
}

func (c *Client) HashExists(key string) (bool, error) {
	_, err := c.GetHash(key)

	if err != nil {
		// Does not exist
		return false, nil
	}

	return true, nil
}

func (c *Client) GetHashValue(key string) ([]KeyPair, error) {
	var keypairs = make([]KeyPair, 0)

	resp, err := c.GetHash(key)

	// Errors on hash lookups *have* to bubble up
	if err != nil {
		return keypairs, err
	}

	for k, v := range resp.Data {
		log.WithFields(log.Fields{
			"key":    key,
			"field":  k,
			"source": v.Source,
		}).Debug("Got hash key")
		keypairs = append(keypairs, KeyPair{k, v.Value, v.Source})
	}
	return keypairs, err
}

func (c *Client) GetStringValueDebug(key string, v ...string) (ValueSource, error) {
	var hasDefault bool
	hasDefault = false
	defaultValue := ValueSource{"", "__DEFAULT__"}
	if len(v) > 0 {
		hasDefault = true
		defaultValue.Value = v[0]
	}

	resp, err := c.GetString(key)
	switch {
	case err != nil && hasDefault:
		// got an error, but we have a default value
		log.WithFields(log.Fields{
			"key":    key,
			"source": "DEFAULT",
		}).Debug("Got string val")
		return defaultValue, nil
	case err != nil:
		log.WithFields(log.Fields{
			"key": key,
		}).Errorf("Cannot get string key and no default value provided: %s", err)
		return defaultValue, err
	default:
		log.WithFields(log.Fields{
			"key":    key,
			"source": resp.Data.Source,
		}).Debug("Got string val")
		return resp.Data, err
	}
}

func (c *Client) StringExists(key string) (bool, error) {
	_, err := c.GetString(key)

	if err != nil {
		// Does not exist
		return false, nil
	}

	return true, nil
}

func (c *Client) GetStringValue(key string, v ...string) (string, error) {
	var hasDefault bool
	hasDefault = false
	defaultValue := ""
	if len(v) > 0 {
		hasDefault = true
		defaultValue = v[0]
	}

	resp, err := c.GetString(key)
	switch {
	case err != nil && hasDefault:
		// got an error, but we have a default value
		log.WithFields(log.Fields{
			"key":    key,
			"source": "DEFAULT",
		}).Debug("Got string val")
		return defaultValue, nil
	case err != nil:
		log.WithFields(log.Fields{
			"key": key,
		}).Errorf("Cannot get string key and no default value provided: %s", err)
		return "", err
	default:
		log.WithFields(log.Fields{
			"key":    key,
			"source": resp.Data.Source,
		}).Debug("Got string val")
		return resp.Data.Value, err
	}
}
