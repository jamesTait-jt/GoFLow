package config

import "errors"

func ValidateDeploymentEnvironment(c Config) error {
	if c.Kubernetes != nil && c.Docker != nil {
		return errors.New("error cannot have both docker and kubernetes configurations")
	}

	return nil
}
