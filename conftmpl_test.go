package confclient

import (
	"testing"
)

func TestMakeRequest(t *testing.T) {
	c := InitiateClient("http://localhost:8080/")

	body, err := c.GETRequestJSON("/string/string")
	if err != nil {
		t.Logf("Error: %s", err)
		t.Fail()
	}
	t.Logf("Body: %s", body)

	body, err = c.GETRequestJSON("/hash/db_policy")
	if err != nil {
		t.Logf("Error: %s", err)
		t.Fail()
	}
	t.Logf("Body: %s", body)

	body, err = c.GETRequestJSON("/list/arraytest")
	if err != nil {
		t.Logf("Error: %s", err)
		t.Fail()
	}
	t.Logf("Body: %s", body)
}

func TestGetStringValue(t *testing.T) {
	c := InitiateClient("http://localhost:8080/")

	resp, err := c.GetStringValue("db_policy/name", "")
	if err != nil {
		t.Logf("Error: %s", err)
		t.Fail()
	}

	t.Logf("Got response: '%s'", resp)
}

func TestGetHash(t *testing.T) {
	c := InitiateClient("http://localhost:8080/")

	resp, err := c.GetHashValue("db_policy")
	if err != nil {
		t.Logf("Error: %s", err)
		t.Fail()
	}

	for _, entry := range resp {
		t.Logf("Got response: %s => %s", entry.Key, entry.Value)
	}
}

func TestGetList(t *testing.T) {
	c := InitiateClient("http://localhost:8080/")

	resp, err := c.GetListValue("arraytest")
	if err != nil {
		t.Logf("Error: %s", err)
		t.Fail()
	}

	for idx, entry := range resp {
		t.Logf("Got response: %d => %s", idx, entry)
	}
}
