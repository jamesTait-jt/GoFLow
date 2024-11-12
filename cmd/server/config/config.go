package config

import (
	"flag"
	"fmt"
)

var defaultBrokerType = "redis"

var supportedBrokerTypes = []string{"redis"}

type Config struct {
	BrokerType string
	BrokerAddr string
}

func LoadConfigFromFlags() *Config {
	c := &Config{}

	enumFlag(&c.BrokerType, "broker-type", supportedBrokerTypes, "Type of task broker (e.g. 'redis')")
	flag.StringVar(&c.BrokerAddr, "broker-addr", "", "Broker address (e.g., Redis address)")

	flag.Parse()

	return c
}

func enumFlag(target *string, name string, allowed []string, usage string) {
	flag.Func(name, usage, func(flagValue string) error {
		if flagValue == "" {
			*target = defaultBrokerType

			return nil
		}

		for _, allowedValue := range allowed {
			if flagValue == allowedValue {
				*target = flagValue
				return nil
			}
		}

		return fmt.Errorf("must be one of %v", allowed)
	})
}
