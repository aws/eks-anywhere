package templater

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/filewriter"
)

type Templater struct {
	writer filewriter.FileWriter
}

func New(writer filewriter.FileWriter) *Templater {
	return &Templater{
		writer: writer,
	}
}

func (t *Templater) WriteToFile(templateContent string, data interface{}, fileName string, f ...filewriter.FileOptionsFunc) (filePath string, err error) {
	bytes, err := Execute(templateContent, data)
	if err != nil {
		return "", err
	}
	writtenFilePath, err := t.writer.Write(fileName, bytes, f...)
	if err != nil {
		return "", fmt.Errorf("writing template file: %v", err)
	}

	return writtenFilePath, nil
}

func (t *Templater) WriteBytesToFile(content []byte, fileName string, f ...filewriter.FileOptionsFunc) (filePath string, err error) {
	writtenFilePath, err := t.writer.Write(fileName, content, f...)
	if err != nil {
		return "", fmt.Errorf("writing template file: %v", err)
	}

	return writtenFilePath, nil
}

func Execute(templateContent string, data interface{}) ([]byte, error) {
	// Apply sprig functions for easy templating.
	// See https://masterminds.github.io/sprig/ for a list of available functions.
	fns := sprig.TxtFuncMap()
	for k, v := range map[string]any{
		"stringsJoin": strings.Join,
		"toYaml":      toYAML,
	} {
		fns[k] = v
	}

	tpl := template.New("").Funcs(fns)
	tpl, err := tpl.Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %v", err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("substituting values for template: %v", err)
	}

	return buf.Bytes(), nil
}

func toYAML(v any) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}
