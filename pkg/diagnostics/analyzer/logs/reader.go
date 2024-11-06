package logs

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/pkg/diagnostics/analyzer"
)

type FSReader struct {
	fs fs.FS
}

func NewFSReader(fs fs.FS) *FSReader {
	return &FSReader{fs: fs}
}

func NewFSReaderForFolder(folder string) *FSReader {
	return NewFSReader(os.DirFS(folder))
}

func (r FSReader) LogsFromDeployment(name, namespace string, filters ...analyzer.LogFilter) ([]string, error) {
	var logs []string

	pods, err := fs.ReadDir(r.fs, namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "looking for logs in namespace %s", namespace)
	}

	for _, pod := range pods {
		if !pod.IsDir() {
			continue
		}

		if !strings.HasPrefix(pod.Name(), name) {
			continue
		}

		containers, err := fs.ReadDir(r.fs, filepath.Join(namespace, pod.Name()))
		if err != nil {
			return nil, errors.Wrapf(err, "looking for containers in pod %s", pod.Name())
		}

		if len(containers) == 0 {
			continue
		}

		if len(containers) > 1 {
			return nil, errors.Errorf("multiple containers in pod %s, not supported", pod.Name())
		}

		containerLogs := containers[0]

		file, err := r.fs.Open(filepath.Join(namespace, pod.Name(), containerLogs.Name()))
		if err != nil {
			return nil, errors.Wrapf(err, "opening log file %s", containerLogs.Name())
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
	scan:
		for scanner.Scan() {
			line := scanner.Text()
			for _, filter := range filters {
				if !filter(line) {
					continue scan
				}
			}
			logs = append(logs, line)
		}

		if err := scanner.Err(); err != nil {
			return nil, errors.Wrapf(err, "reading logs from %s", containerLogs.Name())
		}
	}

	return logs, nil
}
