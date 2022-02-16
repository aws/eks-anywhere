package networkutils

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type IPPool []string

func NewIPPool() IPPool {
	return IPPool{}
}

func NewIPPoolFromString(fromString string) IPPool {
	return IPPool(strings.Split(fromString, ","))
}

func NewIPPoolFromEnv(ipPoolEnvVar string) (IPPool, error) {
	value, ok := os.LookupEnv(ipPoolEnvVar)
	if !ok {
		return NewIPPool(), fmt.Errorf("%s environment ip pool does not exist", ipPoolEnvVar)
	}
	if value != "" {
		return NewIPPoolFromString(value), nil
	}
	return NewIPPool(), nil
}

func (ipPool *IPPool) ToString() string {
	return strings.Join(*ipPool, ",")
}

func (ipPool *IPPool) IsEmpty() bool {
	return len(*ipPool) == 0
}

func (ipPool *IPPool) AddIP(ip string) {
	*ipPool = append(*ipPool, ip)
}

func (ipPool *IPPool) PopIP() (string, error) {
	if ipPool.IsEmpty() {
		return "", errors.New("ip pool is empty")
	} else {
		index := len(*ipPool) - 1
		ip := (*ipPool)[index]
		*ipPool = (*ipPool)[:index]
		return ip, nil
	}
}

func (ipPool *IPPool) ToEnvVar(envVarName string) error {
	s := ipPool.ToString()
	err := os.Setenv(envVarName, s)
	if err != nil {
		return fmt.Errorf("failed to set the ip pool env var %s to value %s", envVarName, s)
	}
	return nil
}
