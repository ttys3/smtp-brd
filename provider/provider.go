package provider

import (
	"fmt"
	"time"
)

const DefaultEmailTimeout = 10 * time.Second

var providers = make(map[string]Factory)

type Factory func () Sender

func registerFactory(provider string, fac Factory) {
	providers[provider] = fac
}

func GetFactory(name string) (fac Factory, err error) {
	if sndr, ok := providers[name]; ok {
		return sndr, nil
	}
	return nil, fmt.Errorf("provider %s not found", name)
}