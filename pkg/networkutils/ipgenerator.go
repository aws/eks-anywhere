package networkutils

type IPGenerator interface {
	GenerateUniqueIP(cidrBlock string) (string, error)
	IsIPUnique(ip string) bool
}
