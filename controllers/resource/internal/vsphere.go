package internal

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func GetVSphereCredValues(credSecret *corev1.Secret) (map[string]string, error) {
	usernameBytes, ok := credSecret.Data["username"]
	if !ok {
		return nil, fmt.Errorf("unable to retrieve username from secret")
	}
	passwordBytes, ok := credSecret.Data["password"]
	if !ok {
		return nil, fmt.Errorf("unable to retrieve password from secret")
	}
	usernameCSIBytes, ok := credSecret.Data["usernameCSI"]
	if !ok {
		usernameCSIBytes = usernameBytes
	}
	passwordCSIBytes, ok := credSecret.Data["passwordCSI"]
	if !ok {
		passwordCSIBytes = passwordBytes
	}
	usernameCPBytes, ok := credSecret.Data["usernameCP"]
	if !ok {
		usernameCPBytes = usernameBytes
	}
	passwordCPBytes, ok := credSecret.Data["passwordCP"]
	if !ok {
		passwordCPBytes = passwordBytes
	}

	values := map[string]string{}
	values["eksaVsphereUsername"] = string(usernameBytes)
	values["eksaVspherePassword"] = string(passwordBytes)
	values["eksaCSIUsername"] = string(usernameCSIBytes)
	values["eksaCSIPassword"] = string(passwordCSIBytes)
	values["eksaCloudProviderUsername"] = string(usernameCPBytes)
	values["eksaCloudProviderPassword"] = string(passwordCPBytes)
	return values, nil
}
