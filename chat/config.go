package chat

import (
	"encoding/json"
	"os"
)

type Config struct {
	Address     string
	WebRoot     string
	Endpoint    string
	HistorySize uint
	Redis       string
}

func LoadConfig(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	c := new(Config)
	if err := json.NewDecoder(f).Decode(c); err != nil {
		return nil, err
	}

	return c, nil
}
