package servicerunner

import (
	"fmt"
	"time"

	"ddns-dnspod/dnspod" // Assuming module path allows this

	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
)

// Program implements service.Interface
type Program struct {
	logger        *logrus.Logger
	ticker        *time.Ticker
	quit          chan struct{}
	secretID      string
	secretKey     string
	domain        string
	recordIDIPv4  int64
	recordIDIPv6  int64
	subDomainIPv4 string
	subDomainIPv6 string
}

// NewProgram creates a new Program instance.
func NewProgram(logger *logrus.Logger, secretID, secretKey, domain string, recordIDIPv4, recordIDIPv6 int64, subDomainIPv4, subDomainIPv6 string) *Program {
	return &Program{
		logger:        logger,
		secretID:      secretID,
		secretKey:     secretKey,
		domain:        domain,
		recordIDIPv4:  recordIDIPv4,
		recordIDIPv6:  recordIDIPv6,
		subDomainIPv4: subDomainIPv4,
		subDomainIPv6: subDomainIPv6,
	}
}

// Start is called when the service is started.
func (p *Program) Start(s service.Service) error {
	p.logger.Info("Service starting...")
	if p.secretID == "" || p.secretKey == "" || p.domain == "" || (p.recordIDIPv4 == 0 && p.recordIDIPv6 == 0) {
		errMsg := "Critical configuration missing (SecretID, SecretKey, Domain, or at least one RecordID for IPv4/IPv6). Service cannot start effectively."
		p.logger.Error(errMsg)
		// Optionally, return an error to prevent the service from starting if config is invalid
		return fmt.Errorf(errMsg)
	}
	if p.recordIDIPv4 == 0 {
		p.logger.Warn("RecordID for IPv4 is not set. IPv4 DDNS updates will be skipped.")
	}
	if p.recordIDIPv6 == 0 {
		p.logger.Warn("RecordID for IPv6 is not set. IPv6 DDNS updates will be skipped.")
	}

	p.quit = make(chan struct{})

	// Initial run
	p.logger.Info("Performing initial DNS update...")
	dnspod.UpdateAndModifyRecords(p.secretID, p.secretKey, p.domain, p.recordIDIPv4, p.recordIDIPv6, p.subDomainIPv4, p.subDomainIPv6, p.logger)

	p.ticker = time.NewTicker(5 * time.Minute)
	go func() {
		p.logger.Info("Background DNS update goroutine started.")
		for {
			select {
			case <-p.ticker.C:
				p.logger.Info("Scheduled DNS update triggered by ticker.")
				dnspod.UpdateAndModifyRecords(p.secretID, p.secretKey, p.domain, p.recordIDIPv4, p.recordIDIPv6, p.subDomainIPv4, p.subDomainIPv6, p.logger)
			case <-p.quit:
				p.ticker.Stop()
				p.logger.Info("Ticker stopped, background goroutine exiting.")
				return
			}
		}
	}()
	p.logger.Info("Service started successfully.")
	return nil
}

// Stop is called when the service is stopped.
func (p *Program) Stop(s service.Service) error {
	p.logger.Info("Service stopping...")
	if p.quit != nil {
		close(p.quit)
	}
	// Add any other cleanup logic here
	p.logger.Info("Service stopped.")
	return nil
}

// GetDomain returns the configured domain.
func (p *Program) GetDomain() string {
	return p.domain
}

// GetRecordIDIPv4 returns the configured IPv4 Record ID.
func (p *Program) GetRecordIDIPv4() int64 {
	return p.recordIDIPv4
}

// GetRecordIDIPv6 returns the configured IPv6 Record ID.
func (p *Program) GetRecordIDIPv6() int64 {
	return p.recordIDIPv6
}

// GetSubDomainIPv4 returns the configured IPv4 SubDomain.
func (p *Program) GetSubDomainIPv4() string {
	if p.subDomainIPv4 == "" {
		return "@" // Default to root if not specified
	}
	return p.subDomainIPv4
}

// GetSubDomainIPv6 returns the configured IPv6 SubDomain.
func (p *Program) GetSubDomainIPv6() string {
	if p.subDomainIPv6 == "" {
		return "@" // Default to root if not specified
	}
	return p.subDomainIPv6
}

// GetSecretID returns the configured Secret ID.
func (p *Program) GetSecretID() string {
	return p.secretID
}
