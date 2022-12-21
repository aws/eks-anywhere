package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const defaultInstanceIDBlockSize = 500

// TerminateEc2Instances terminates EC2 instances by calling the AWS API.
func TerminateEc2Instances(session *session.Session, instanceIDs []*string) error {
	service := ec2.New(session)
	for _, instanceIDChunk := range makeChunks(instanceIDs, defaultInstanceIDBlockSize) {
		input := &ec2.TerminateInstancesInput{
			InstanceIds: instanceIDChunk,
		}

		if _, err := service.TerminateInstances(input); err != nil {
			return fmt.Errorf("terminating EC2 instances: %v", err)
		}
	}

	return nil
}

func makeChunks[T any](elements []T, chunkSize int) [][]T {
	elementsSize := len(elements)
	if elementsSize == 0 {
		return [][]T{}
	}
	chunksSize := (elementsSize + chunkSize - 1) / chunkSize
	chunks := make([][]T, 0, chunksSize)
	start := 0
	for i := 0; i < chunksSize-1; i++ {
		end := start + chunkSize
		chunks = append(chunks, elements[start:end])
		start = end
	}
	chunks = append(chunks, elements[start:elementsSize])

	return chunks
}
