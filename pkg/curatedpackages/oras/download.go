package oras

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/logger"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type BundleDownloader struct {
	dstFolder string
	log       logr.Logger
}

// NewBundleDownloader returns a new BundleDownloader.
func NewBundleDownloader(log logr.Logger, dstFolder string) *BundleDownloader {
	return &BundleDownloader{
		log:       log,
		dstFolder: dstFolder,
	}
}

func (bd *BundleDownloader) Download(ctx context.Context, bundles *releasev1.Bundles) {
	artifacts := ReadFilesFromBundles(bundles)
	for _, a := range UniqueCharts(artifacts) {
		data, err := curatedpackages.PullLatestBundle(ctx, bd.log, a)
		if err != nil {
			fmt.Printf("unable to download bundle %v \n", err)
			continue
		}
		bundleName := strings.Replace(filepath.Base(a), ":", "-", 1)
		err = writeToFile(bd.dstFolder, bundleName, data)
		if err != nil {
			fmt.Printf("unable to write to file %v \n", err)
		}
	}
}

func UniqueCharts(charts []string) []string {
	keys := make(map[string]bool)
	var list []string

	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range charts {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func writeToFile(dir string, packageName string, content []byte) error {
	file := filepath.Join(dir, packageName) + ".yaml"
	if err := os.WriteFile(file, content, 0o640); err != nil {
		return fmt.Errorf("unable to write to the file: %s %v", file, err)
	}
	return nil
}

func ReadFilesFromBundles(bundles *releasev1.Bundles) []string {
	var files []string
	for _, vb := range bundles.Spec.VersionsBundles {
		file, err := curatedpackages.GetPackageBundleRef(vb)
		if err != nil {
			logger.Info("Warning: Failed parsing package bundle reference", "error", err)
			continue
		}
		files = append(files, file)
	}
	return files
}
