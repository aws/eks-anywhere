package aws

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	metadataServiceIP = "169.254.169.254"
	tokenTTLSeconds   = 21600
)

var (
	// IMDSv2TokenURL contains the full url to the aws IMDSv2 session token endpoint.
	IMDSv2TokenURL = fmt.Sprintf("http://%s/latest/api/token", metadataServiceIP)
	// IMDSv2PublicIPv4URL contains the full url to the aws IMDSv2 public ipv4 endpoint.
	IMDSv2PublicIPv4URL = fmt.Sprintf("http://%s/latest/meta-data/public-ipv4", metadataServiceIP)
)

// GenerateIMDSv2SessionToken generates a session token that can be used to call aws IMDSv2.
func (c *Client) GenerateIMDSv2SessionToken(url string) (string, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("creating http POST request for generating IMDSv2 session token: %v", err)
	}

	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", strconv.Itoa(tokenTTLSeconds))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("generating IMDSv2 session token from IMDSv2 url [%s]: %v", url, err)
	}
	defer resp.Body.Close()

	token, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body from IMDSv2 url [%s]: %v", url, err)
	}

	return string(token), nil
}

// InstanceIP fetches the instance public ipv4 with a session token through aws IMDSv2.
func (c *Client) InstanceIP(url, token string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("creating http GET request for fetching instance IP through IMDSv2: %v", err)
	}

	req.Header.Set("X-aws-ec2-metadata-token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching instance IP from IMDSv2 url [%s]: %v", url, err)
	}
	defer resp.Body.Close()

	instanceIP, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body from IMDSv2 url [%s]: %v", url, err)
	}

	return string(instanceIP), nil
}
