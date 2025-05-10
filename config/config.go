package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

// AppConfig defines the configuration structure.
// RecordIDs are kept as strings here to match TOML and env, conversion happens later.
type AppConfig struct {
	SecretID      string `toml:"DNSPOD_SECRET_ID"`
	SecretKey     string `toml:"DNSPOD_SECRET_KEY"`
	Domain        string `toml:"DNSPOD_DOMAIN"`
	RecordIDIPv4  string `toml:"DNSPOD_RECORDID_IPV4"`  // Kept as string for initial loading
	RecordIDIPv6  string `toml:"DNSPOD_RECORDID_IPV6"`  // Kept as string for initial loading
	SubDomainIPv4 string `toml:"DNSPOD_SUBDOMAIN_IPV4"` // New field for IPv4 subdomain
	SubDomainIPv6 string `toml:"DNSPOD_SUBDOMAIN_IPV6"` // New field for IPv6 subdomain
}

// Load loads configuration from the specified file path and environment variables.
// Environment variables override file configurations.
func Load(configFileArg string, logger *logrus.Logger) (AppConfig, error) {
	var cfg AppConfig
	configPath := configFileArg

	if configPath == "" {
		exePath, err := os.Executable()
		if err == nil {
			configPath = filepath.Join(filepath.Dir(exePath), "config.toml")
		} else {
			logger.Warnf("警告: 无法确定可执行文件路径以查找 config.toml: %v。将依赖环境变量。", err)
		}
	}

	if configPath != "" {
		if _, statErr := os.Stat(configPath); statErr == nil {
			if _, decodeErr := toml.DecodeFile(configPath, &cfg); decodeErr != nil {
				logger.Warnf("警告: 读取配置文件 %s 失败: %v。将依赖环境变量。", configPath, decodeErr)
			} else {
				logger.Infof("成功从 %s 加载配置。", configPath)
			}
		} else {
			if configFileArg != "" { // Only log "not found" if a specific path was given
				logger.Infof("信息: 指定的配置文件 %s 未找到。将依赖环境变量。", configPath)
			} else if !os.IsNotExist(statErr) { // Some other error occurred
				logger.Warnf("信息: 检查配置文件 %s 时出错: %v。将依赖环境变量。", configPath, statErr)
			} else { // Default path and file does not exist
				logger.Infof("信息: 未在默认位置找到配置文件 %s。将依赖环境变量。", configPath)
			}
		}
	}

	// Override with environment variables
	if envSecretID := os.Getenv("DNSPOD_SECRET_ID"); envSecretID != "" {
		cfg.SecretID = envSecretID
	}
	if envSecretKey := os.Getenv("DNSPOD_SECRET_KEY"); envSecretKey != "" {
		cfg.SecretKey = envSecretKey
	}
	if envDomain := os.Getenv("DNSPOD_DOMAIN"); envDomain != "" {
		cfg.Domain = envDomain
	}
	if envRecordIdIPv4 := os.Getenv("DNSPOD_RECORDID_IPV4"); envRecordIdIPv4 != "" {
		cfg.RecordIDIPv4 = envRecordIdIPv4
	}
	if envRecordIdIPv6 := os.Getenv("DNSPOD_RECORDID_IPV6"); envRecordIdIPv6 != "" {
		cfg.RecordIDIPv6 = envRecordIdIPv6
	}
	if envSubDomainIPv4 := os.Getenv("DNSPOD_SUBDOMAIN_IPV4"); envSubDomainIPv4 != "" {
		cfg.SubDomainIPv4 = envSubDomainIPv4
	}
	if envSubDomainIPv6 := os.Getenv("DNSPOD_SUBDOMAIN_IPV6"); envSubDomainIPv6 != "" {
		cfg.SubDomainIPv6 = envSubDomainIPv6
	}

	// Basic validation
	if cfg.SecretID == "" || cfg.SecretKey == "" || cfg.Domain == "" || (cfg.RecordIDIPv4 == "" && cfg.RecordIDIPv6 == "") {
		errMsg := "警告: DNSPOD_SECRET_ID, DNSPOD_SECRET_KEY, DNSPOD_DOMAIN, 或至少一个 DNSPOD_RECORDID_IPV4/DNSPOD_RECORDID_IPV6 未在配置文件或环境变量中完全设置。"
		logger.Warn(errMsg)
	}

	return cfg, nil
}
