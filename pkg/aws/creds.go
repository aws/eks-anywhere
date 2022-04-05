package aws

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"

	"github.com/aws/eks-anywhere/pkg/validations"
)

const (
	EksaAwsCredentialsFileKey = "EKSA_AWS_CREDENTIALS_FILE"
	EksaAwsCABundlesFileKey   = "EKSA_AWS_CA_BUNDLES_FILE"
)

func AwsCredentialsFile() (filePath string, err error) {
	return validateFileFromEnv(EksaAwsCredentialsFileKey)
}

func AwsCABundlesFile() (filePath string, err error) {
	return validateFileFromEnv(EksaAwsCABundlesFileKey)
}

func validateFileFromEnv(envKey string) (filePath string, err error) {
	filePath, ok := os.LookupEnv(envKey)
	if !ok || len(filePath) <= 0 {
		return "", fmt.Errorf("env %s is not set or is empty", envKey)
	}

	fileExists := validations.FileExists(filePath)
	if !fileExists {
		return "", fmt.Errorf("file %s does not exist", filePath)
	}
	return filePath, nil
}

func EncodeFileFromEnv(envKey string) (string, error) {
	filePath, err := validateFileFromEnv(envKey)
	if err != nil {
		return "", err
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to read file due to: %v", err)
	}

	return base64.StdEncoding.EncodeToString(content), nil
}

func ParseDeviceIPs(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ips := []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := subtractProfileName(scanner.Text())
		if net.ParseIP(line) != nil {
			ips = append(ips, line)
		}
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no ip address profile found in file %s", filePath)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ips, nil
}

func subtractProfileName(input string) string {
	re := regexp.MustCompile(`^\[((.*?))\]$`)
	if match := re.FindStringSubmatch(input); match != nil {
		return match[1]
	}
	return ""
}
