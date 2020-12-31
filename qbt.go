package qbt

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
)

// Client Client
type Client struct {
	http   *http.Client
	config Config
}

// New New
func New(url string, username string, password string) (*Client, error) {
	client := newClient(url, username, password)
	if !client.isLogin() {
		err := client.login(username, password)
		return client, err
	}
	return client, nil
}

// List 获取下载列表
func (client *Client) List() (Sync, error) {
	connected := client.GetConnectionStatus()
	var s Sync
	if !connected {
		return s, errors.New("无法连接到服务器，请检查网络和配置")
	}
	params := make(map[string]string)
	params["rid"] = "0"

	resp, err := client.get("api/v2/sync/maindata", params)
	if err != nil {
		return s, err
	}
	json.NewDecoder(resp.Body).Decode(&s)
	return s, nil
}

// Download 下载种子
func (client *Client) Download(url string) error {
	connected := client.GetConnectionStatus()
	if !connected {
		return errors.New("无法连接到服务器，请检查网络和配置")
	}

	credentials := make(map[string]string)
	credentials["urls"] = url
	credentials["savepath"] = "/downloads/"
	credentials["autoTMM"] = "false"
	credentials["paused"] = "false"
	credentials["root_folder"] = "true"

	resp, err := client.post("api/v2/torrents/add", credentials)
	if err != nil {
		return err
	} else if resp.Status != "200 OK" { // check for correct status code
		return errors.New(resp.Status)
	}
	return nil
}

// GetConnectionStatus 获取连接状态
func (client *Client) GetConnectionStatus() bool {
	if !client.isLogin() {
		err := client.login(client.config.Username, client.config.Password)
		if err != nil {
			return false
		}
		return true
	}
	return true
}

//NewClient creates a new client connection to qbittorrent
func newClient(url string, username string, password string) *Client {
	client := &Client{}
	// ensure url ends with "/"
	if url[len(url)-1:] != "/" {
		url += "/"
	}
	client.config = Config{
		URL:      url,
		Username: username,
		Password: password,
	}
	// create cookie jar
	jar, _ := cookiejar.New(nil)
	client.http = &http.Client{
		Jar: jar,
	}
	return client
}

//Login logs you in to the qbittorrent client
//returns the current authentication status
func (client *Client) login(username string, password string) error {
	credentials := make(map[string]string)
	credentials["username"] = username
	credentials["password"] = password

	resp, err := client.post("api/v2/auth/login", credentials)
	if err != nil {
		return err
	} else if resp.Status != "200 OK" { // check for correct status code
		return errors.New("couldnt log in")
	}
	return nil
}

// IsLogin IsLogin
func (client *Client) isLogin() bool {
	resp, err := client.get("api/v2/app/version", make(map[string]string))
	if err != nil {
		return false
	} else if resp.Status != "200 OK" {
		return false
	}
	return true
}

//post will perform a POST request with no content-type specified
func (client *Client) post(endpoint string, opts map[string]string) (*http.Response, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	for k, v := range opts {
		_ = bodyWriter.WriteField(k, v)
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	req, err := http.NewRequest("POST", client.config.URL+endpoint, bodyBuf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", contentType)

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, nil

}

//get will perform a GET request with no parameters
func (client *Client) get(endpoint string, opts map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", client.config.URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	// add optional parameters that the user wants
	if opts != nil {
		query := req.URL.Query()
		for k, v := range opts {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
