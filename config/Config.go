package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"sqlow/database"
	"sqlow/helpers"
	"strings"
)

type Config struct {
	Engine   string            `yaml:"engine"`
	Host     string            `yaml:"host"`
	Port     string            `yaml:"port"`
	Schema   string            `yaml:"schema"`
	Username string            `yaml:"username"`
	Password string            `yaml:"-"`
	Options  map[string]string `yaml:"options"`
}

func (c *Config) WithOverrides(engine, host, port, schema, username, password, options interface{}) Config {
	config := Config{
		c.Engine,
		c.Host,
		c.Port,
		c.Schema,
		c.Username,
		c.Password,
		c.Options,
	}
	if engine != nil {
		config.Engine = engine.(string)
	}
	if host != nil {
		config.Host = host.(string)
	}
	if port != nil {
		config.Port = port.(string)
	}
	if schema != nil {
		config.Schema = schema.(string)
	}
	if username != nil {
		config.Username = username.(string)
	}
	if password != nil {
		config.Password = password.(string)
	}
	if options != nil {
		cliOptions := strings.Split(options.(string), ",")
		for _, optionPair := range cliOptions {
			splitPair := strings.Split(optionPair, ":")
			c.Options[splitPair[0]] = splitPair[1]
		}
	}
	log.Println("Loaded CLI options.")
	return config
}

func (c *Config) MakeDriver() database.Driver {
	log.Printf("Making %s driver...\n", c.Engine)
	var options []string
	for key, value := range c.Options {
		options = append(options, fmt.Sprintf("%s=%s", key, value))
	}
	return database.Driver{
		Engine:   c.Engine,
		Host:     c.Host,
		Port:     c.Port,
		Schema:   c.Schema,
		Username: c.Username,
		Password: c.Password,
		Options:  options,
	}
}

func FromYAML(f string) *Config {
	b, err := os.ReadFile(f)
	helpers.CheckWarn(err)
	if err != nil {
		log.Println("Loaded default options.")
		return &Config{
			"postgres",
			"localhost",
			"5432",
			"postgres",
			"postgres",
			"",
			map[string]string{},
		}
	}
	var c Config
	err = yaml.Unmarshal(b, &c)
	helpers.CheckError(err)

	log.Printf("Loaded config from %s.\n", f)
	return &c
}
