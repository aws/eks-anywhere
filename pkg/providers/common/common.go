package common

import (
	_ "embed"
)

//go:embed config/audit-policy.yaml
var auditPolicy string

func GetAuditPolicy() string {
	return auditPolicy
}
