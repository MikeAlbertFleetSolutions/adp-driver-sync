package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/MikeAlbertFleetSolutions/adp-driver-sync/adp"
	"github.com/MikeAlbertFleetSolutions/adp-driver-sync/config"
	"github.com/MikeAlbertFleetSolutions/adp-driver-sync/mikealbert"
)

var (
	buildnum string
)

func main() {
	// show file & location, date & time
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// command line app
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "\nUsage of %s build %s\n", os.Args[0], buildnum)
		flag.PrintDefaults()
	}

	// process command line
	var configFile string
	flag.StringVar(&configFile, "config", "", "Configuration file")
	flag.Parse()

	if len(configFile) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// read config
	err := config.FromFile(configFile)
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	// create adp client
	ac, err := adp.NewClient(config.Adp.ClientId, config.Adp.ClientSecret, config.Adp.BaseURL, config.Adp.CertFile, config.Adp.KeyFile)
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	// create mike albert client
	mac, err := mikealbert.NewClient(config.MikeAlbert.ClientId, config.MikeAlbert.ClientSecret, config.MikeAlbert.Endpoint)
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	// get employees from ADP
	drivers, err := ac.GetDriverHomeAddresses()
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	log.Printf("Found %d drivers from ADP", len(drivers))

	// sync each driver to mike albert
	updated := 0
	unchanged := 0
	notFound := 0
	skipped := 0
	errors := 0

	for _, d := range drivers {
		// Mike Albert stores employee numbers without leading zeros
		employeeNumber := strings.TrimLeft(d.EmployeeNumber, "0")

		// find the driver in mike albert by employee number
		maDrivers, err := mac.FindDrivers(employeeNumber)
		if err != nil {
			log.Printf("ERROR finding driver %s in Mike Albert: %+v", employeeNumber, err)
			errors++
			continue
		}

		if len(maDrivers) == 0 {
			notFound++
			continue
		}

		// update each matching driver in mike albert
		for _, maDriver := range maDrivers {
			// Compare current MA address with ADP address — only PATCH if different
			newZip := d.ZIPCode
			if len(newZip) > 5 {
				newZip = newZip[:5]
			}
			currentZip := maDriver.Address.PostCode
			if len(currentZip) > 5 {
				currentZip = currentZip[:5]
			}

			if strings.EqualFold(strings.TrimSpace(maDriver.Address.Address1), strings.TrimSpace(d.Address1)) &&
				strings.EqualFold(strings.TrimSpace(maDriver.Address.Address2), strings.TrimSpace(d.Address2)) &&
				currentZip == newZip {
				unchanged++
				continue
			}

			log.Printf("  Updating DriverId %d (%s): '%s' -> '%s', '%s' -> '%s', '%s' -> '%s'",
				*maDriver.DriverId, employeeNumber,
				maDriver.Address.Address1, d.Address1,
				maDriver.Address.Address2, d.Address2,
				maDriver.Address.PostCode, d.ZIPCode)

			_, err = mac.UpdateDriver(*maDriver.DriverId, d.Address1, d.Address2, d.ZIPCode)
			if err != nil {
				if strings.Contains(err.Error(), "multiple vehicles allocated") {
					log.Printf("  WARN: DriverId %d has multiple vehicles - skipping address update", *maDriver.DriverId)
					skipped++
				} else {
					log.Printf("  ERROR updating DriverId %d for EmployeeNumber %s: %+v", *maDriver.DriverId, employeeNumber, err)
					errors++
				}
				continue
			}

			log.Printf("  SUCCESS: Updated DriverId %d", *maDriver.DriverId)
			updated++
		}
	}

	log.Printf("=== SYNC COMPLETE ===")
	log.Printf("  Total ADP drivers:   %d", len(drivers))
	log.Printf("  Updated:             %d", updated)
	log.Printf("  Unchanged:           %d", unchanged)
	log.Printf("  Not found in MA:     %d", notFound)
	log.Printf("  Skipped (multi-veh): %d", skipped)
	log.Printf("  Errors:              %d", errors)
}
