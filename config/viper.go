package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var v *viper.Viper

type BrdConfig struct {
	Provider string `toml:"provider"`

	Addr string `toml:"addr"`
	Port string `toml:"port"`

	CertFile string `toml:"cert"`
	KeyFile string `toml:"key"`

	AuthUsername string `toml:"user"`
	AuthPassword string `toml:"secret"`

	TLS bool  `toml:"tls"`
	Debug bool `toml:"debug"`
}

func init() {
	v = viper.NewWithOptions(viper.EnvKeyReplacer(strings.NewReplacer(".", "_")))
	v.SetConfigName("config") //do not include extension in order to auto lookup
	v.SetConfigType("toml")
	v.AddConfigPath(".")           // optionally look for config in the working directory
	v.AddConfigPath("/etc/brd")    // path to look for the config file in
	v.SetEnvPrefix("BRD")
	v.AutomaticEnv()
	//v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.Set("Verbose", true)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; just warn
			fmt.Println("no config file found")
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("error parse config file: %s \n", err))
		}
	} else {
		fmt.Println("config loaded: " + v.ConfigFileUsed())
	}
}

func V() *viper.Viper {
	return v
}