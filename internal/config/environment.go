package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	EnvironmentLocal       = Environment("local")
	EnvironmentDevelopment = Environment("development")
	EnvironmentStaging     = Environment("staging")
	EnvironmentProduction  = Environment("production")
)

var environmentMap = map[string]Environment{
	EnvironmentLocal.String():       EnvironmentLocal,
	EnvironmentDevelopment.String(): EnvironmentDevelopment,
	EnvironmentStaging.String():     EnvironmentStaging,
	EnvironmentProduction.String():  EnvironmentProduction,
}

type Environment string

func (f Environment) IsOneOf(envs ...Environment) bool {
	return slices.Contains(envs, f)
}

func (f Environment) String() string {
	return string(f)
}

func (f *Environment) SetValue(s string) error {
	if s == "" {
		return errors.New("environment field value can't be empty")
	}

	env, ok := environmentMap[strings.ToLower(s)]
	if !ok {
		return fmt.Errorf("unknown environment: %s", s)
	}

	*f = env
	return nil
}
