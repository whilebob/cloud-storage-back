package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

var AppConfig Config

type Config struct {
	MySQL *MySQLConfig `yaml:"mysql"`
	Redis *RedisConfig `yaml:"redis"`
	Minio *MinioConfig `yaml:"minio"`
	HOST  string       `yaml:"host"`
	PORT  int          `yaml:"port"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DB       int    `yaml:"db"`
	Password string `yaml:"password"`
}

type MinioConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

// LoadConfig 加载配置文件
func LoadConfig() error {

	getwd, err := os.Getwd()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(getwd + "/config/config.yaml")
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &AppConfig)
	if err != nil {
		return err
	}

	return nil
}
