package dnspod

import (
	"ddns-dnspod/ipfetcher" // Assuming module path allows this
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspodapi "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323" // Alias to avoid conflict
)

// ModifyRecord updates a DNS record on DNSPod using the specific SDK.
func ModifyRecord(domain string, recordId int64, value string, secretID string, secretKey string, recordType string, subDomain string, logger *logrus.Logger) {
	credential := common.NewCredential(secretID, secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"

	client, errClient := dnspodapi.NewClient(credential, "", cpf)
	if errClient != nil {
		logger.Errorf("Failed to create DNSPod client: %v", errClient)
		return
	}

	request := dnspodapi.NewModifyRecordRequest()

	request.Domain = common.StringPtr(domain)
	request.RecordType = common.StringPtr(recordType)
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(value)
	request.RecordId = common.Uint64Ptr(uint64(recordId)) // Convert int64 to uint64

	// Use provided subdomain, default to "@" if empty
	actualSubDomain := subDomain
	if actualSubDomain == "" {
		actualSubDomain = "@"
	}
	request.SubDomain = common.StringPtr(actualSubDomain)

	request.TTL = common.Uint64Ptr(600) // Default TTL

	logger.Debugf("Modifying DNSPod record: Domain=%s, Type=%s, Line=%s, Value=%s, RecordID=%d, SubDomain=%s, TTL=%d",
		*request.Domain, *request.RecordType, *request.RecordLine, *request.Value, *request.RecordId, *request.SubDomain, *request.TTL)

	response, err := client.ModifyRecord(request)
	if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok {
		logger.Errorf("DNSPod API error occurred: Code=%s, Message=%s, RequestId=%s", sdkErr.GetCode(), sdkErr.GetMessage(), sdkErr.GetRequestId())
		return
	}
	if err != nil {
		logger.Errorf("Failed to invoke ModifyRecord API: %v", err)
		return
	}

	responseBody := response.ToJsonString()
	logger.Infof("ModifyRecord API Response for %s (%s): %s", domain, recordType, responseBody)
}

// UpdateAndModifyRecords fetches current IP addresses and updates DNS records.
func UpdateAndModifyRecords(secretID, secretKey, domain string, recordIDIPv4, recordIDIPv6 int64, subDomainIPv4, subDomainIPv6 string, logger *logrus.Logger) {
	// The 'domain' parameter here is expected to be the main domain (e.g., "example.com").

	logger.Info("Fetching current IPv4 address...")
	ipv4, err := ipfetcher.GetCurrentIP(ipfetcher.IPv4URL, logger)
	if err != nil {
		logger.Errorf("Error getting IPv4 address: %v", err)
	} else {
		logger.Infof("Current IPv4 Address: %s", ipv4)
		if recordIDIPv4 != 0 {
			ModifyRecord(domain, recordIDIPv4, ipv4, secretID, secretKey, "A", subDomainIPv4, logger)
		} else {
			logger.Warn("RecordID for IPv4 is not set. Skipping A record update.")
		}
	}

	logger.Info("Fetching current IPv6 address...")
	ipv6, err := ipfetcher.GetCurrentIP(ipfetcher.IPv6URL, logger)
	if err != nil {
		logger.Errorf("Error getting IPv6 address: %v", err)
	} else {
		logger.Infof("Current IPv6 Address: %s", ipv6)
		if recordIDIPv6 != 0 {
			ModifyRecord(domain, recordIDIPv6, ipv6, secretID, secretKey, "AAAA", subDomainIPv6, logger)
		} else {
			logger.Warn("RecordID for IPv6 is not set. Skipping AAAA record update.")
		}
	}
}

// Helper to parse domain and subdomain, if needed in the future.
func parseDomain(fullDomain string) (subDomain, mainDomain string) {
	parts := strings.Split(fullDomain, ".")
	if len(parts) > 2 { // e.g., sub.example.com
		subDomain = strings.Join(parts[:len(parts)-2], ".")
		mainDomain = strings.Join(parts[len(parts)-2:], ".")
	} else { // e.g., example.com or single part
		subDomain = "@" // Or handle as per API requirements for root domain
		mainDomain = fullDomain
	}
	return
}
