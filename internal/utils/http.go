package utils

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/nikhiljohn10/uagplugin/logger"
)

// IsRepoPublic checks GitHub repo visibility with a bounded timeout and context.
// Caller should pass a context with deadline; a default 5s client timeout also applies.
func IsRepoPublic(ctx context.Context, apiURL string, token string) bool {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
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
