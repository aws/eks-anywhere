package executables

import "fmt"

type cmkCommandArgs func(*[]string)

func newCmkCommand(command string) []string {
	return []string{command}
}

func applyCmkArgs(params *[]string, args ...cmkCommandArgs) {
	for _, arg := range args {
		arg(params)
	}
}

func appendArgs(new ...string) cmkCommandArgs {
	return func(args *[]string) {
		*args = append(*args, new...)
	}
}

func withCloudStackDomainId(domainId string) cmkCommandArgs {
	return appendArgs(fmt.Sprintf("domainid=\"%s\"", domainId))
}

func withCloudStackAccount(account string) cmkCommandArgs {
	return appendArgs(fmt.Sprintf("account=\"%s\"", account))
}

func withCloudStackZoneId(zoneId string) cmkCommandArgs {
	return appendArgs(fmt.Sprintf("zoneid=\"%s\"", zoneId))
}

func withCloudStackNetworkType(networkType string) cmkCommandArgs {
	return appendArgs(fmt.Sprintf("type=\"%s\"", networkType))
}

func withCloudStackId(id string) cmkCommandArgs {
	return appendArgs(fmt.Sprintf("id=\"%s\"", id))
}

func withCloudStackName(name string) cmkCommandArgs {
	return appendArgs(fmt.Sprintf("name=\"%s\"", name))
}
