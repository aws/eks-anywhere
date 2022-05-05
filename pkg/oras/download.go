package oras

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type BundleDownloader struct {
	dstFolder string
}

func NewBundleDownloader(dstFolder string) *BundleDownloader {
	return &BundleDownloader{
		dstFolder: dstFolder,
	}
}

func (bd *BundleDownloader) SaveManifests(ctx context.Context, bundles *releasev1.Bundles) {
	artifacts := ReadFilesFromBundles(bundles)
	for _, a := range UniqueCharts(artifacts) {
		data, err := curatedpackages.Pull(ctx, a)
		if err != nil {
			fmt.Printf("unable to download bundle %v", err)
			continue
		}
		bundleName := strings.Replace(filepath.Base(a), ":", "-", 1)
		err = writeToFile(bd.dstFolder, bundleName, data)
		if err != nil {
			fmt.Printf("unable to write to file %v", err)
		}
	}
}

func UniqueCharts(charts []string) []string {
	c := types.SliceToLookup(charts).ToSlice()
	// TODO: maybe optimize this, avoiding the sort and just following the same order as the original slice
	sort.Strings(c)
	return c
}

func writeToFile(dir string, packageName string, content []byte) error {
	file := filepath.Join(dir, packageName) + ".yaml"
	if err := os.WriteFile(file, content, 0o644); err != nil {
		return fmt.Errorf("unable to write to the file: %s %v", file, err)
	}
	return nil
}

func ReadFilesFromBundles(bundles *releasev1.Bundles) []string {
	var files []string
	for _, vb := range bundles.Spec.VersionsBundles {
		files = append(files, curatedpackages.GetPackageBundleRef(vb))
	}
	return files
}
