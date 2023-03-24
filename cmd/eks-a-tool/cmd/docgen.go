package cmd

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	anywhere "github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd"
)

const fmTemplate = `---
title: "%s"
linkTitle: "%s"
---

`

var cmdDocPath string

var docgenCmd = &cobra.Command{
	Use:    "docgen",
	Short:  "Generate the documentation for the CLI commands",
	Long:   "Use eks-a-tool docgen to auto generate CLI commands documentation",
	Hidden: true,
	RunE:   docgenCmdRun,
}

func init() {
	docgenCmd.Flags().StringVar(&cmdDocPath, "path", "./docs/content/en/docs/reference/eksctl", "Path to write the generated documentation to")
	rootCmd.AddCommand(docgenCmd)
}

func docgenCmdRun(_ *cobra.Command, _ []string) error {
	anywhereRootCmd := anywhere.RootCmd()
	anywhereRootCmd.DisableAutoGenTag = true
	if err := doc.GenMarkdownTreeCustom(anywhereRootCmd, cmdDocPath, filePrepender, linkHandler); err != nil {
		return fmt.Errorf("error generating markdown doc from eksctl-anywhere root cmd: %v", err)
	}
	return nil
}

func filePrepender(filename string) string {
	name := filepath.Base(filename)
	base := strings.TrimSuffix(name, path.Ext(name))
	title := strings.Replace(base, "_", " ", -1)
	return fmt.Sprintf(fmTemplate, title, title)
}

func linkHandler(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	base = strings.Replace(base, "(", "", -1)
	base = strings.Replace(base, ")", "", -1)
	return "../" + strings.ToLower(base) + "/"
}
