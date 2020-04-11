package provider

import (
	"fmt"
	"time"
)

const DefaultEmailTimeout = 10 * time.Second

type providerPool map[string]Factory

var providers = make(providerPool)

type Factory func() Sender

func registerFactory(provider string, fac Factory) {
	providers[provider] = fac
}

func GetFactory(name string) (fac Factory, err error) {
	if sndr, ok := providers[name]; ok {
		return sndr, nil
	}
	return nil, fmt.Errorf("provider %s not found", name)
}

func AvailableProviders() []string {
	var poolNames []string
	for name := range providers {
		poolNames = append(poolNames, name)
	}
	return poolNames
}

func (p *providerPool) String() string {
	return fmt.Sprintf("%+v", AvailableProviders())
}