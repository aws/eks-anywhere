package curatedpackages

import (
	"strings"
)

const (
	Download            = "download"
	Import              = "import"
	packageProdLocation = "783794618700.dkr.ecr.us-west-2.amazonaws.com"
	packageDevLocation  = "857151390494.dkr.ecr.us-west-2.amazonaws.com"
	publicECR           = "public.ecr.aws/eks-anywhere"
)

type PackageReaderType string

func NewReader(s string) PackageReaderType {
	return PackageReaderType(strings.ToLower(strings.TrimSpace(s)))
}

func (r *PackageReaderType) GetRegistry(uri string) string {
	switch *r {
	case Download:
		if strings.Contains(uri, publicECR) {
			return packageProdLocation
		}
		return packageDevLocation
	default:
		return GetRegistry(uri)
	}
}
