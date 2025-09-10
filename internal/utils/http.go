package utils

import (
	"io"
	"net/http"
	"regexp"

	"github.com/nikhiljohn10/uagplugin/logger"
)

func IsRepoPublic(apiURL string, token string) bool {
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logger.Warn("Could not create request for repo privacy: %v", err)
		return false
	}
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("Could not check repo privacy, assuming private: %v", err)
		return false
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("Could not read repo privacy response, assuming private: %v", err)
		return false
	}
	r, _ := regexp.Compile(`\"private\":false`)
	return r.MatchString(string(body))
}
