package artifacts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

const (
	realmKey   = "realm="
	serviceKey = "service="
	scopeKey   = "scope="
)

type CheckImageExistence struct {
	ImageUri   string
	AuthHeader string
}

type tokenResponse struct {
	Token string `json:"token"`
}

func (d CheckImageExistence) Run(ctx context.Context) error {
	registry, repository, tag, error := splitImageUri(d.ImageUri)
	if error != nil {
		return error
	}
	requestUrl := fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, repository, tag)

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return errors.Cause(err)
	}
	req.Header.Add("Authorization", d.AuthHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("requested image not found")
	} else if resp.StatusCode == http.StatusUnauthorized && len(d.AuthHeader) == 0 {
		splits := strings.Split(resp.Header.Get("www-authenticate"), ",")
		var realm, service, scope string
		for _, split := range splits {
			if strings.Contains(split, realmKey) {
				startIndex := strings.Index(split, realmKey) + len(realmKey)
				realm = strings.Trim(split[startIndex:], "\"")
			} else if strings.Contains(split, serviceKey) {
				startIndex := strings.Index(split, serviceKey) + len(serviceKey)
				service = strings.Trim(split[startIndex:], "\"")
			} else if strings.Contains(split, scopeKey) {
				startIndex := strings.Index(split, scopeKey) + len(scopeKey)
				scope = strings.Trim(split[startIndex:], "\"")
			}
		}
		token, err := getRegistryToken(realm, service, scope)
		if err != nil {
			return err
		}
		d.AuthHeader = "Bearer " + token
		return d.Run(ctx)
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unknown response: %s", resp.Status)
	}
	return nil
}

func getRegistryToken(realm, service, scope string) (string, error) {
	requestUrl := fmt.Sprintf("%s?service=\"%s\"&scope=\"%s\"", realm, service, scope)

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return "", errors.Cause(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to token from %s", requestUrl)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	tokenResp := tokenResponse{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	return tokenResp.Token, nil
}

func splitImageUri(imageUri string) (string, string, string, error) {
	indexOfSlash := strings.Index(imageUri, "/")
	if indexOfSlash < 0 {
		return "", "", "", errors.Errorf("Invalid URI: %s", imageUri)
	}
	registry := imageUri[:indexOfSlash]
	imageUriSplit := strings.Split(imageUri[len(registry)+1:], ":")
	if len(imageUriSplit) < 2 {
		return "", "", "", errors.Errorf("Invalid URI: %s", imageUri)
	}
	repository := strings.Replace(imageUriSplit[0], registry+"/", "", -1)
	tag := imageUriSplit[1]

	return registry, repository, tag, nil
}
