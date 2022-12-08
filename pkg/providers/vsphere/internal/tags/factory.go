package tags

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Factory struct {
	client GovcClient
}

type GovcClient interface {
	ListTags(ctx context.Context) ([]executables.Tag, error)
	CreateTag(ctx context.Context, tag, category string) error
	AddTag(ctx context.Context, path, tag string) error
	ListCategories(ctx context.Context) ([]string, error)
	CreateCategoryForVM(ctx context.Context, name string) error
}

func NewFactory(client GovcClient) *Factory {
	return &Factory{client}
}

func (f *Factory) TagTemplate(ctx context.Context, templatePath string, tagsByCategory map[string][]string) error {
	logger.V(2).Info("Tagging template", "template", templatePath)
	categories, err := f.client.ListCategories(ctx)
	if err != nil {
		return fmt.Errorf("failed listing vsphere categories: %v", err)
	}

	tags, err := f.client.ListTags(ctx)
	if err != nil {
		return fmt.Errorf("failed listing vsphere tags: %v", err)
	}

	tagNames := make([]string, 0, len(tags))
	for _, t := range tags {
		tagNames = append(tagNames, t.Name)
	}

	categoriesLookup := types.SliceToLookup(categories)
	tagsLookup := types.SliceToLookup(tagNames)
	for category, tags := range tagsByCategory {
		if !categoriesLookup.IsPresent(category) {
			logger.V(3).Info("Creating category", "category", category)
			if err = f.client.CreateCategoryForVM(ctx, category); err != nil {
				return fmt.Errorf("failed creating category for tags: %v", err)
			}
		}

		for _, tag := range tags {
			if !tagsLookup.IsPresent(tag) {
				logger.V(3).Info("Creating tag", "tag", tag, "category", category)
				if err = f.client.CreateTag(ctx, tag, category); err != nil {
					return fmt.Errorf("failed creating tag before tagging template: %v", err)
				}
			}

			logger.V(3).Info("Adding tag to template", "tag", tag, "template", templatePath)
			if err = f.client.AddTag(ctx, templatePath, tag); err != nil {
				return fmt.Errorf("failed tagging template: %v", err)
			}
		}
	}

	return nil
}
