package application

import (
	"flag"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents configuration for the application
type Config struct {
	Postgres Postgres `yaml:"postgres"`
	IFaces   IFaces   `yaml:"ifaces"`
	Logger   Logger   `yaml:"logger"`
	Kafka    Kafka    `yaml:"kafka"`
}

// Postgres represents configuration data for establishing connection
type Postgres struct {
	DSN string `yaml:"connection-string"`
}

// IFaces contains ports on services
type IFaces struct {
	RESTPort    string `yaml:"rest_port"`
	MSGPort     string `yaml:"msg_listener_port"`
	AUTHAddress string `yaml:"auth_server_address"`
}

// Logger has values for the logger
type Logger struct {
	Level string `yaml:"level"`
}

// Kafka keeps values to connect
type Kafka struct {
	Server  string `yaml:"server"`
	Topic   string `yaml:"topic"`
	GroupID string `yaml:"group_id"`
}

func getConfig() *Config {
	path := flag.String("c", "./configs/config.yaml", "set path to config yaml-file")
	flag.Parse()

	log.Printf("config file, %s", *path)

	f, err := os.Open(*path)
	if err != nil {
		log.Fatalf("cannot open %s config file: %v", *path, err)
	}
	defer f.Close()

	return readConfigFile(f)
}

// read parses yaml file to get application Config
func readConfigFile(r io.Reader) *Config {

	cfg := &Config{}
	d := yaml.NewDecoder(r)
	if err := d.Decode(cfg); err != nil {
		log.Fatalf("cannot parse config %v", err)
	}
	return cfg
}
