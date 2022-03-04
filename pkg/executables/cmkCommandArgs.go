package executables

import "fmt"

type CmkCommandArgs func(*[]string)

func newCmkCommand(command string) []string {
	return []string{command}
}

func ApplyCmkArgs(params *[]string, args ...CmkCommandArgs) {
	for _, arg := range args {
		arg(params)
	}
}

func appendArgs(new ...string) CmkCommandArgs {
	return func(args *[]string) {
		*args = append(*args, new...)
	}
}
func WithCloudStackDomainId(domainId string) CmkCommandArgs {
	return appendArgs(fmt.Sprintf("domainid=\"%s\"", domainId))
}

func WithCloudStackAccount(account string) CmkCommandArgs {
	return appendArgs(fmt.Sprintf("account=\"%s\"", account))
}

func WithCloudStackZoneId(zoneId string) CmkCommandArgs {
	return appendArgs(fmt.Sprintf("zoneid=\"%s\"", zoneId))
}

func WithCloudStackNetworkType(networkType string) CmkCommandArgs {
	return appendArgs(fmt.Sprintf("type=\"%s\"", networkType))
}

func WithCloudStackId(id string) CmkCommandArgs {
	return appendArgs(fmt.Sprintf("id=\"%s\"", id))
}

func WithCloudStackName(name string) CmkCommandArgs {
	return appendArgs(fmt.Sprintf("name=\"%s\"", name))
}
