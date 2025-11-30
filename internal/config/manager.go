package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type IConfig interface {
	Server() ServerConfig
	HttpClient() HttpClientConfig
	Scheduler() SchedulerConfig
	WebhookConfig() WebhookConfig
	Database() DatabaseConfig
	Redis() RedisConfig
}

var GlobalConfig IConfig

type config struct {
	cfg Config
}

func NewConfig() IConfig {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalf(err.Error())
	}
	GlobalConfig = &config{cfg: cfg}
	return GlobalConfig
}

func (c *config) Server() ServerConfig {
	return c.cfg.Server
}

func (c *config) HttpClient() HttpClientConfig {
	return c.cfg.HttpClient
}

func (c *config) Scheduler() SchedulerConfig {
	return c.cfg.Scheduler
}

func (c *config) WebhookConfig() WebhookConfig {
	return c.cfg.WebhookConfig
}

func (c *config) Database() DatabaseConfig {
	return c.cfg.Database
}

func (c *config) Redis() RedisConfig {
	return c.cfg.Redis
}
