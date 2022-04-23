package features

import (
	"os"
	"strings"
	"sync"
)

var globalFeatures = newFeatures()

type features struct {
	cache     *mutexMap
	gates     map[string]string
	initGates sync.Once
}

func newFeatures() *features {
	return &features{
		cache: newMutexMap(),
		gates: map[string]string{},
	}
}

func (f *features) feedGates(featureGates []string) {
	f.initGates.Do(func() {
		for _, gate := range featureGates {
			pairs := strings.SplitN(gate, "=", 2)
			if len(pairs) != 2 {
				continue
			}

			f.gates[pairs[0]] = pairs[1]
		}
	})
}

func (f *features) isActiveForEnvVar(envVar string) func() bool {
	return func() bool {
		active, ok := f.cache.load(envVar)
		if !ok {
			active = os.Getenv(envVar) == "true"
			f.cache.store(envVar, active)
		}

		return active
	}
}

func (f *features) isActiveForEnvVarOrGate(envVar, gateName string) func() bool {
	return func() bool {
		active, ok := f.cache.load(envVar)
		if !ok {
			value, present := os.LookupEnv(envVar)
			if !present {
				value = f.gates[gateName]
			}

			active = value == "true"
			f.cache.store(envVar, active)
		}

		return active
	}
}

func (f *features) clearCache() {
	f.cache.clear()
}
