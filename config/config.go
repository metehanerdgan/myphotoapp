package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Config, uygulama konfigürasyonunu temsil eder.
type Config struct {
	Vision struct {
		CredentialsFile string `yaml:"credentialsFile"`
	} `yaml:"vision"`
	Kafka struct {
		Broker string `yaml:"broker"`
	} `yaml:"kafka"`
}

// LoadConfig, belirtilen YAML dosyasından konfigürasyonu yükler.
func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Konfigürasyon dosyası açılamadı: %v", err)
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	var config Config
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Konfigürasyon dosyası çözümlenemedi: %v", err)
		return nil, err
	}

	return &config, nil
}
