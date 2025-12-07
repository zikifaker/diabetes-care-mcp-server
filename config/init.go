package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

var Cfg Config

type Config struct {
	Server struct {
		Port     string `yaml:"port"`
		LogLevel string `yaml:"log_level"`
	}
	DB struct {
		Neo4j DBConfig `yaml:"neo4j"`
	} `yaml:"db"`
	Milvus struct {
		Endpoint string `yaml:"endpoint"`
		APIKey   string `yaml:"api_key"`
	} `yaml:"milvus"`
	Model struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"model"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func init() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %v", err))
	}

	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		panic(fmt.Sprintf("Failed to parse config: %v", err))
	}
}
