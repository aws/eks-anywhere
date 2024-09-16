package e2e

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/aws/eks-anywhere/pkg/logger"
)

var svc *cloudwatch.CloudWatch

func init() {
	if s, err := session.NewSession(); err == nil {
		svc = cloudwatch.New(s)
	} else {
		fmt.Println("Cannot create CloudWatch service", err)
	}
}

func putInstanceTestResultMetrics(r instanceTestsResults) {
	if svc == nil {
		logger.Info("Cannot publish metrics as cloudwatch service was not initialized")
		return
	}

	logger.Info("Publishing instance test result metrics")
	// Note 0 metrics are emitted for the purpose of aggregation. For example, when the succeededCount metrics are [0, 1, 0, 1], we can calculate the success rate as 2 / 4 = 50%. However, when 0 are excluded, the metrics becomes [1, 1], and you would not be able to calculate the success rate from that series.
	erroredCount, failedCount, succeededCount := 0, 0, 0
	if r.err != nil {
		erroredCount = 1
	} else if !r.testCommandResult.Successful() {
		failedCount = 1
	} else {
		succeededCount = 1
	}

	data := &cloudwatch.MetricDatum{
		Unit: aws.String("Count"),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("Provider"),
				Value: aws.String(getProviderName(r.conf.regex)),
			},
			{
				Name:  aws.String("BranchName"),
				Value: aws.String(r.conf.branchName),
			},
		},
		Timestamp: aws.Time(time.Now()),
	}
	putMetric(data, "ErroredInstanceTests", erroredCount)
	putMetric(data, "FailedInstanceTests", failedCount)
	putMetric(data, "SucceededInstanceTests", succeededCount)

	// TODO: publish time metrics
	logger.Info("Test instance metrics published")
}

func getProviderName(testRe string) string {
	providerRe := regexp.MustCompile(`Test((?i:vsphere)|(?i:cloudstack)|(?i:snow)|(?i:docker)|(?i:nutanix)|(?i:tinkerbell))`)
	provider := []byte("Unknown")
	t := providerRe.FindSubmatch([]byte(testRe))
	if len(t) > 1 {
		provider = t[1]
	}
	return strings.ToLower(string(provider))
}

func putMetric(data *cloudwatch.MetricDatum, metricName string, value int) {
	data.MetricName = aws.String(metricName)
	data.Value = aws.Float64(float64(value))

	if _, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace:  aws.String("EksaE2ETests"),
		MetricData: []*cloudwatch.MetricDatum{data},
	}); err != nil {
		logger.Error(err, "Cannot put metrics to cloudwatch")
	} else {
		logger.Info("Instance test result metrics published")
	}
}
