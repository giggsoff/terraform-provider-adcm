package client

import "encoding/json"

type results struct {
	Results interface{} `json:"results"`
}

func unwrapResults(body []byte, placeholder interface{}) error {
	var obj results
	obj.Results = placeholder
	err := json.Unmarshal(body, &obj)
	if err != nil {
		return err
	}
	return nil
}
