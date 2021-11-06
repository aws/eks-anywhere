package features

import (
	"os"
)

const (
	AwsIamAuthenticatorEnvVar = "AWS_IAM_AUTHENTICATOR"
)

var cache = newMutexMap()

type Feature struct {
	Name     string
	IsActive func() bool
}

func IsActive(feature Feature) bool {
	return feature.IsActive()
}

func isActiveForEnvVar(envVar string) func() bool {
	return func() bool {
		active, ok := cache.load(envVar)
		if !ok {
			active = os.Getenv(envVar) == "true"
			cache.store(envVar, active)
		}

		return active
	}
}

func AwsIamAuthenticator() Feature {
	return Feature{
		Name:     "aws-iam-authenticator identity provider",
		IsActive: isActiveForEnvVar(AwsIamAuthenticatorEnvVar),
	}
}
