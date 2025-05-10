package main

import (
	"flag"
	"os"
	"strconv"

	"ddns-dnspod/config"
	"ddns-dnspod/logger"
	"ddns-dnspod/servicerunner" // Renamed package for clarity

	"github.com/kardianos/service"
)

const serviceName = "DDNSDNSPODService"
const serviceDisplayName = "DDNS DNSPOD Service"
const serviceDescription = "Dynamic DNS service for DNSPOD."

func main() {
	// Initialize logger early. isInteractive is determined by kardianos/service.
	// We pass service.Interactive() to logger.Init.
	// The serviceName constant is used for the log file name.
	logger.Init(service.Interactive(), "ddns-server") // "ddns-server" will be part of log file name

	log := logger.L() // Get the initialized logger instance

	svcConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceDisplayName,
		Description: serviceDescription,
		// Dependencies:    []string{"After=network.target"}, // Example for systemd
	}

	// Application-specific flags
	appFlagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError) // ContinueOnError to handle service commands
	configFile := appFlagSet.String("c", "", "Path to the config.toml file")

	// The Program instance will be created after config is loaded.
	// service.New requires a service.Interface, so we'll create Program later.
	// For now, we prepare a placeholder or defer its creation.
	var prg *servicerunner.Program

	// Create service controller. We pass a nil Program initially,
	// as it needs configuration that's loaded after flag parsing.
	// This is a bit tricky. service.New expects a non-nil service.Interface.
	// We will create a temporary Program for service commands that don't run the core logic.
	// For the actual run, we'll re-initialize `prg` with full config.

	// A minimal program for service management commands (install/remove)
	// that don't need full config.
	minimalPrg := servicerunner.NewProgram(log, "", "", "", 0, 0, "", "")
	s, err := service.New(minimalPrg, svcConfig)
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	// Handle service control arguments first.
	// These arguments (install, remove, start, stop) are typically exclusive.
	if len(os.Args) > 1 {
		serviceAction := os.Args[1]
		switch serviceAction {
		case "install":
			err := s.Install()
			if err != nil {
				log.Fatalf("Failed to install service: %v", err)
			}
			log.Info("Service installed successfully.")
			return
		case "remove":
			err := s.Uninstall()
			if err != nil {
				log.Fatalf("Failed to remove service: %v", err)
			}
			log.Info("Service removed successfully.")
			return
		case "start": // OS service manager calls this, or user manually.
			// s.Run() will eventually call prg.Start()
			// If called directly like `myapp.exe start`, it might just mean "run now".
			// The service library handles the actual service start.
			// We let s.Run() handle this.
			log.Info("Attempting to start service via service manager or direct run...")
			// No direct s.Start() call here, s.Run() handles it.
		case "stop": // OS service manager calls this.
			// Similar to start, service library handles this.
			log.Info("Attempting to stop service via service manager...")
			// No direct s.Stop() call here.
		default:
			// If not a known service command, try to parse app flags.
			if err := appFlagSet.Parse(os.Args[1:]); err != nil {
				if err == flag.ErrHelp {
					// appFlagSet.PrintDefaults() // Or custom help message
					os.Exit(0)
				}
				log.Warnf("Failed to parse application flags: %v. Use -h for help.", err)
				// os.Exit(1) // Exit if flag parsing fails for non-service commands
			}
		}
	} else {
		// No arguments, or only os.Args[0]. Try to parse for -c if any.
		// This handles running the app directly without service commands.
		if err := appFlagSet.Parse(os.Args[1:]); err != nil {
			if err == flag.ErrHelp {
				os.Exit(0)
			}
			// log.Warnf("Failed to parse application flags on direct run: %v", err)
			// Allow running without flags if config is via env or default file.
		}
	}

	log.Info("DDNS Service Application Logic Starting/Resuming...")

	appCfg, err := config.Load(*configFile, log)
	if err != nil {
		// config.Load logs warnings but doesn't return fatal errors for missing files.
		// We might want to be stricter here if essential configs are absent.
		log.Errorf("Failed to load configuration: %v", err)
		// If running as a service, this might be an issue.
		// If interactive, the user sees the log.
	}

	if appCfg.SecretID == "" || appCfg.SecretKey == "" || appCfg.Domain == "" || (appCfg.RecordIDIPv4 == "" && appCfg.RecordIDIPv6 == "") {
		log.Error("Critical configuration (SecretID, SecretKey, Domain, and at least one of RecordIDIPv4 or RecordIDIPv6) is missing or incomplete.")
		if service.Interactive() {
			log.Info("Please ensure configuration is set via config.toml or environment variables.")
			os.Exit(1) // Exit if interactive and config is bad
		}
		// If not interactive (i.e., running as a service), Program.Start will also log this.
		// The service might fail to start properly.
	}

	var recordIdIPv4Int64 int64
	if appCfg.RecordIDIPv4 != "" {
		recordIdIPv4Int64, err = strconv.ParseInt(appCfg.RecordIDIPv4, 10, 64)
		if err != nil {
			log.Fatalf("无法将 RecordIDIPv4 (%s) 转换为 int64: %v", appCfg.RecordIDIPv4, err)
		}
	} else {
		log.Warn("DNSPOD_RECORDID_IPV4 is not set. IPv4 updates will be skipped if not running as a service and this is the only ID missing.")
	}

	var recordIdIPv6Int64 int64
	if appCfg.RecordIDIPv6 != "" {
		recordIdIPv6Int64, err = strconv.ParseInt(appCfg.RecordIDIPv6, 10, 64)
		if err != nil {
			log.Fatalf("无法将 RecordIDIPv6 (%s) 转换为 int64: %v", appCfg.RecordIDIPv6, err)
		}
	} else {
		log.Warn("DNSPOD_RECORDID_IPV6 is not set. IPv6 updates will be skipped if not running as a service and this is the only ID missing.")
	}

	if recordIdIPv4Int64 == 0 && recordIdIPv6Int64 == 0 && !service.Interactive() {
		// If running as a service and BOTH RecordIDs are effectively zero/unset.
		log.Error("Neither DNSPOD_RECORDID_IPV4 nor DNSPOD_RECORDID_IPV6 is set. Service cannot perform any DNS updates.")
	}

	// Now create the actual Program with loaded configuration
	prg = servicerunner.NewProgram(log, appCfg.SecretID, appCfg.SecretKey, appCfg.Domain, recordIdIPv4Int64, recordIdIPv6Int64, appCfg.SubDomainIPv4, appCfg.SubDomainIPv6)

	// Update the service with the fully configured program
	// This is a common pattern: create service with a placeholder, then update its interface.
	// However, kardianos/service.New takes the interface at creation.
	// So, we need to create a new service instance if we want to pass the fully configured `prg`.
	// Or, ensure `minimalPrg` can be updated, but `service.Interface` is not directly mutable on `s`.
	// The simplest is to re-create `s` if `prg` is different from `minimalPrg`.
	// But since `install`/`remove` exit, for the `run` path, we can create `s` here.

	// If we reached here, it's not an install/remove command.
	// We are either running directly or being started as a service.
	// Create the service with the fully configured program.
	s, err = service.New(prg, svcConfig)
	if err != nil {
		log.Fatalf("Failed to create service with full config: %v", err)
	}

	// Log effective configuration being used by the service runner
	log.Infof("Service will run with DNSPOD_DOMAIN: %s", prg.GetDomain())
	if prg.GetRecordIDIPv4() != 0 {
		log.Infof("DNSPOD_RECORDID_IPV4: %d, SUBDOMAIN_IPV4: %s", prg.GetRecordIDIPv4(), prg.GetSubDomainIPv4())
	} else {
		log.Info("DNSPOD_RECORDID_IPV4: Not set or invalid, IPv4 updates will be skipped.")
	}
	if prg.GetRecordIDIPv6() != 0 {
		log.Infof("DNSPOD_RECORDID_IPV6: %d, SUBDOMAIN_IPV6: %s", prg.GetRecordIDIPv6(), prg.GetSubDomainIPv6())
	} else {
		log.Info("DNSPOD_RECORDID_IPV6: Not set or invalid, IPv6 updates will be skipped.")
	}
	log.Infof("DNSPOD_SECRET_ID is set: %t", prg.GetSecretID() != "")

	err = s.Run()
	if err != nil {
		log.Errorf("Service run failed: %v", err)
	}
}
