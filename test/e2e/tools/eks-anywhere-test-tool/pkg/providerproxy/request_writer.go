package providerproxy

import (
	"bytes"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/aws/eks-anywhere-test-tool/pkg/filewriter"
)

type requestWriter struct {
	filewriter.FileWriter
}

func newRequestWriter(folderPath string) (*requestWriter, error) {
	writer := filewriter.NewWriter(folderPath)
	return &requestWriter{FileWriter: writer}, nil
}

func (w *requestWriter) writeRequest(logs []*cloudwatchlogs.OutputLogEvent) error {
	buf := new(bytes.Buffer)
	for _, log := range logs {
		buf.WriteString(*log.Message + "\n")
	}

	if _, err := w.Write("requests", buf.Bytes(), filewriter.PersistentFile); err != nil {
		return err
	}

	return nil
}
