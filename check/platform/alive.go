package platform

import (
	"net/http"

	"github.com/tao-t356/subs-check/config"
)

func CheckAlive(httpClient *http.Client) (bool, error) {
	resp, err := httpClient.Get(config.Current().AliveTestUrl)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	// 2xx
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, nil
	}
	return false, nil
}
