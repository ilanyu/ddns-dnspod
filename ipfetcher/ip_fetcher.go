package ipfetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

// IPDetails 用于解析来自 ipinfo.app 的JSON响应
type IPDetails struct {
	IP            string `json:"ip"`
	ASN           string `json:"asn"`
	Continent     string `json:"continent"`
	ContinentLong string `json:"continentLong"`
	Flag          string `json:"flag"`
	Country       string `json:"country"`
}

const (
	IPv4URL = "https://ipv4.my.ipinfo.app/api/ipDetails.php"
	IPv6URL = "https://ipv6.my.ipinfo.app/api/ipDetails.php"
)

// GetCurrentIP 从指定的URL获取IP地址
func GetCurrentIP(url string, logger *logrus.Logger) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get IP from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get IP from %s: status code %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	var ipDetails IPDetails
	err = json.Unmarshal(body, &ipDetails)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON response from %s: %w", url, err)
	}

	if ipDetails.IP == "" {
		return "", fmt.Errorf("no IP address found in response from %s", url)
	}
	logger.Debugf("Fetched IP %s from %s", ipDetails.IP, url)
	return ipDetails.IP, nil
}
