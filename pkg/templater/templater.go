package templater

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

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
	temp := template.New("tmpl")
	funcMap := map[string]interface{}{
		"indent": func(spaces int, v string) string {
			pad := strings.Repeat(" ", spaces)
			return pad + strings.Replace(v, "\n", "\n"+pad, -1)
		},
		"stringsJoin": strings.Join,
	}
	temp = temp.Funcs(funcMap)

	temp, err := temp.Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %v", err)
	}

	var buf bytes.Buffer
	err = temp.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("substituting values for template: %v", err)
	}
	return buf.Bytes(), nil
}
