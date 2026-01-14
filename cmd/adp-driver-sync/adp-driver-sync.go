package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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

	// get employees to sync over
	drivers, err := ac.GetDriverHomeAddresses()
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	// update drivers in mike albert
	for _, d := range drivers {
		maDrivers, err := mac.FindDrivers(d.EmployeeNumber)
		if err != nil {
			log.Printf("EmployeeNumber %s: %+v", d.EmployeeNumber, err)
			continue
		}

		for _, maDriver := range maDrivers {
			_, err = mac.UpdateDriver(*maDriver.DriverId, d.Address1, d.Address2, d.ZIPCode)
			if err != nil {
				log.Printf("EmployeeNumber %s: %+v", d.EmployeeNumber, err)
				continue
			}
		}
	}
}
