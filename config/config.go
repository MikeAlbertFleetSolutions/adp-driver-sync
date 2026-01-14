package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	msgMissingField = "required configuration missing %s"

	// unwrapped config values
	Adp        adp
	MikeAlbert mikealbert
)

type configuration struct {
	Adp        adp
	MikeAlbert mikealbert
}

func (c *configuration) validate() error {
	if err := c.Adp.validate(); err != nil {
		return err
	}
	if err := c.MikeAlbert.validate(); err != nil {
		return err
	}
	return nil
}

type adp struct {
	ClientId     string
	ClientSecret string
	BaseURL      string
	CertFile     string
	KeyFile      string
}

func (a *adp) validate() error {
	if len(a.ClientId) == 0 {
		return fmt.Errorf("ADP ClientId is required")
	}
	if len(a.ClientSecret) == 0 {
		return fmt.Errorf("ADP ClientSecret is required")
	}
	if len(a.BaseURL) == 0 {
		return fmt.Errorf("ADP BaseURL is required")
	}
	if len(a.CertFile) == 0 {
		return fmt.Errorf("ADP CertFile is required")
	}
	if len(a.KeyFile) == 0 {
		return fmt.Errorf("ADP KeyFile is required")
	}
	return nil
}

type mikealbert struct {
	ClientId     string
	ClientSecret string
	Endpoint     string
}

func (m *mikealbert) validate() error {
	if len(m.ClientId) == 0 {
		err := fmt.Errorf(msgMissingField, "Mike Albert ClientId")
		log.Printf("%+v", err)
		return err
	}
	if len(m.ClientSecret) == 0 {
		err := fmt.Errorf(msgMissingField, "Mike Albert ClientSecret")
		log.Printf("%+v", err)
		return err
	}
	if len(m.Endpoint) == 0 {
		err := fmt.Errorf(msgMissingField, "Mike Albert Endpoint")
		log.Printf("%+v", err)
		return err
	}

	return nil
}

// FromFile reads the application configuration from file configFile
func FromFile(configFile string) error {
	// read config
	bytes, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	var c configuration
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// validation
	err = c.validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	Adp = c.Adp
	MikeAlbert = c.MikeAlbert

	return nil
}

// Write writes configuration to the file configFile
func Write(configFile string) error {
	// wrap
	c := configuration{
		Adp:        Adp,
		MikeAlbert: MikeAlbert,
	}

	// make sure valid before proceeding
	err := c.validate()
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// create YAML to write
	b, err := yaml.Marshal(c)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	// write out to file
	err = os.WriteFile(configFile, b, 0600)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	return nil
}
