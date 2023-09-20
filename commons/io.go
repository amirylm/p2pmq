package commons

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func ReadConfig(cfgAddr string) (*Config, error) {
	raw, err := os.ReadFile(cfgAddr)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	cfg := Config{}
	if strings.Contains(cfgAddr, ".yaml") || strings.Contains(cfgAddr, ".yml") {
		err = yaml.Unmarshal(raw, &cfg)
	} else {
		err = json.Unmarshal(raw, &cfg)
	}
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %w", err)
	}
	return &cfg, nil
}
