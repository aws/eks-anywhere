package providerProxy

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func VsphereSessionsFilter(logs []*cloudwatchlogs.OutputLogEvent) (outputLogs []*cloudwatchlogs.OutputLogEvent, err error) {
	vsphereSoapSessionType := "SessionManager"
	var sessionLogs []*cloudwatchlogs.OutputLogEvent
	for _, log := range logs {
		if strings.Contains(*log.Message, vsphereSoapSessionType) {
			sessionLogs = append(sessionLogs, log)
		}
	}
	return sessionLogs, nil
}
