package eureka

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type UAAClient struct {
	BaseURL string
	Name    string
	Secret  string
}

type CheckTokenResponse struct {
	Scope    []string `json:"scope"`
	UserID   string   `json:"user_id"`
	UserName string   `json:"user_name"`
}

func (c *UAAClient) GetToken() (string, error) {
	reqURL := fmt.Sprintf("%s/oauth/token", c.BaseURL)
	bodyString := fmt.Sprintf("client_id=%s&grant_type=client_credentials", c.Name)
	request, err := http.NewRequest("POST", reqURL, strings.NewReader(bodyString))
	request.SetBasicAuth(c.Name, c.Secret)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	type getTokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	response := &getTokenResponse{}
	err = c.makeRequest(request, response)
	if err != nil {
		return "", err
	}
	return response.AccessToken, nil
}

func (c *UAAClient) makeRequest(request *http.Request, response interface{}) error {
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("http client: %s", err)
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %s", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad uaa response, code %d, msg %s", resp.StatusCode, string(respBytes))
	}

	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		return fmt.Errorf("unmarshal json: %s", err)
	}
	return nil
}
