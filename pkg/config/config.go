package config

type Config struct {
	Port        int    `yaml:"PORT"`
	Password    string `yaml:"PASSWORD"`
	DstHost     string `yaml:"DST_HOST"`
	DstPort     int    `yaml:"DST_PORT"`
	DstUsername string `yaml:"DST_USERNAME"`
	DstPassword string `yaml:"DST_PASSWORD"`
}
