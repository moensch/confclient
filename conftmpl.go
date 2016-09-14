package confclient

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

func (c *Client) GetListValue(key string) ([]string, error) {
	var strings = make([]string, 0)

	resp, err := c.GetList(key)

	// Errors on list lookups *have* to bubble up
	if err != nil {
		return strings, err
	}

	for _, v := range resp.Data {
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

	return resp.Data, err
}

func (c *Client) GetHashValue(key string) ([]KeyPair, error) {
	var keypairs = make([]KeyPair, 0)

	resp, err := c.GetHash(key)

	// Errors on hash lookups *have* to bubble up
	if err != nil {
		return keypairs, err
	}

	for k, v := range resp.Data {
		keypairs = append(keypairs, KeyPair{k, v.Value, v.Source})
	}
	return keypairs, err
}

func (c *Client) GetStringValueDebug(key string, v ...string) (ValueSource, error) {
	defaultValue := ValueSource{}
	if len(v) > 0 {
		defaultValue.Value = v[0]
	}

	resp, err := c.GetString(key)
	if err != nil {
		// Errors return empty string
		return defaultValue, nil
	}

	if resp.Data.Value == "" {
		return defaultValue, err
	}
	return resp.Data, err
}

func (c *Client) GetStringValue(key string, v ...string) (string, error) {
	defaultValue := ""
	if len(v) > 0 {
		defaultValue = v[0]
	}

	resp, err := c.GetString(key)
	if err != nil {
		// Errors return empty string
		return defaultValue, nil
	}

	if resp.Data.Value == "" {
		return defaultValue, err
	}
	return resp.Data.Value, err
}
