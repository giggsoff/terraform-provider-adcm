package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// SignIn - Get a new token for user
func (c *Client) SignIn() (*AuthResponse, error) {
	if c.Auth.Username == "" || c.Auth.Password == "" {
		return nil, fmt.Errorf("define username and password")
	}

	resp, err := http.PostForm(fmt.Sprintf("%s/api/v1/token/", c.HostURL),
		url.Values{
			"username": {c.Auth.Username},
			"password": {c.Auth.Password},
		})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("getToken: http status: %s", resp.Status)
		body, err := io.ReadAll(io.LimitReader(resp.Body, MaxPostSize))
		if err == nil {
			log.Printf("getToken: POST body: '%s'", body)
		}
		return nil, fmt.Errorf("wrong responce status: %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxPostSize))
	if err != nil {
		return nil, fmt.Errorf("getToken: body read error: %s", err)
	}

	ar := AuthResponse{}
	err = json.Unmarshal(body, &ar)
	if err != nil {
		return nil, err
	}

	return &ar, nil
}
