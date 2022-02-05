package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kelseyhightower/envconfig"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v2"
)

var c = &BrdConfig{}

func init() {
	c.ProviderConfigs = make(map[string]PluginConfig)
}

type BrdConfig struct {
	Provider        string                  `toml:"provider" yaml:"provider"`
	Addr            string                  `toml:"addr" yaml:"addr"`
	Port            string                  `toml:"port" yaml:"port"`
	CertFile        string                  `toml:"cert" yaml:"cert_file"`
	KeyFile         string                  `toml:"key" yaml:"key_file"`
	AuthUsername    string                  `toml:"user" yaml:"auth_username"`
	AuthPassword    string                  `toml:"secret" yaml:"auth_password"`
	TLS             bool                    `toml:"tls" yaml:"tls"`
	Debug           bool                    `toml:"debug" yaml:"debug"`
	ProviderConfigs map[string]PluginConfig `toml:"provider_configs" yaml:"provider_configs"`
}

type PluginConfig map[string]string

func (p PluginConfig) GetString(key string, defaultValue ...string) string {
	var ret string
	if p == nil {
		return ""
	}
	if val, ok := p[key]; ok {
		return val
	} else if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ret
}

func (p PluginConfig) GetInt(key string, defaultValue ...int) int {
	var ret int
	if p == nil {
		return 0
	}

	if val, ok := p[key]; ok {
		val, _ := strconv.Atoi(val)
		return val
	} else if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ret
}

func (p PluginConfig) GetBool(key string, defaultValue ...bool) bool {
	if p == nil {
		return false
	}
	if val, ok := p[key]; ok {
		val, _ := strconv.ParseBool(val)
		return val
	} else if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

func Cfg() *BrdConfig {
	return c
}

func Load(configFile string) error {
	var err error
	var content []byte
	ext := filepath.Ext(configFile)
	switch ext {
	case ".yaml", ".yml":
		content, err = os.ReadFile(configFile)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(content, c)
		if err != nil {
			return err
		}
	case ".toml":
		content, err = os.ReadFile(configFile)
		if err != nil {
			return err
		}
		err = toml.Unmarshal(content, c)
		if err != nil {
			return err
		}
	default:
		err = envconfig.Process("brd", c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *BrdConfig) Dump() {
	out, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}

func (c *BrdConfig) ProviderConfig(provider string) PluginConfig {
	return c.ProviderConfigs[provider]
}

func (c *BrdConfig) InitDefaultProviderConfig(provider string, cfg PluginConfig) {
	c.ProviderConfigs[provider] = cfg
}
