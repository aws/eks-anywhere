package resource

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceUpdater interface {
	CreateResource(ctx context.Context, obj *unstructured.Unstructured, dryRun bool) error
	UpdateTemplate(template *unstructured.Unstructured, values map[string]interface{}) (hasDiff bool, err error)
	ApplyTemplate(ctx context.Context, template *unstructured.Unstructured, values map[string]interface{}, dryRun bool) error
	ForceApplyTemplate(ctx context.Context, template *unstructured.Unstructured, dryRun bool) error
	ApplyUpdatedTemplate(ctx context.Context, template *unstructured.Unstructured, dryRun bool) error
	ApplyPatch(ctx context.Context, obj client.Object, dryRun bool) error
}

type capiResourceUpdater struct {
	client client.Client
	Log    logr.Logger
}

func NewCAPIResourceUpdater(client client.Client, log logr.Logger) *capiResourceUpdater {
	return &capiResourceUpdater{
		client: client,
		Log:    log,
	}
}

func (u *capiResourceUpdater) ApplyPatch(ctx context.Context, obj client.Object, dryRun bool) error {
	dryRunStage := []string{}
	if dryRun {
		dryRunStage = []string{"All"}
	}
	err := u.client.Patch(ctx, obj, client.Merge, &client.PatchOptions{FieldManager: "eks-controller", DryRun: dryRunStage})
	if err != nil {
		return err
	}
	return nil
}

func (u *capiResourceUpdater) CreateResource(ctx context.Context, obj *unstructured.Unstructured, dryRun bool) error {
	dryRunStage := []string{}
	if dryRun {
		dryRunStage = []string{"All"}
	}
	obj.SetResourceVersion("")
	err := u.client.Create(ctx, obj, &client.CreateOptions{FieldManager: "eks-controller", DryRun: dryRunStage})
	if err != nil {
		return err
	}
	return nil
}

func (u *capiResourceUpdater) UpdateTemplate(template *unstructured.Unstructured, values map[string]interface{}) (hasDiff bool, err error) {
	originalTemplate := template.DeepCopy()
	for k, v := range values {
		path := strings.Split(k, ",")
		err := validatePath(template, path)
		if err != nil {
			return false, err
		}
		err = unstructured.SetNestedField(template.Object, v, path...)
		if err != nil {
			return false, err
		}
	}
	return !reflect.DeepEqual(originalTemplate.Object, template.Object), err
}

func validatePath(template *unstructured.Unstructured, path []string) error {
	_, exists, err := unstructured.NestedFieldNoCopy(template.Object, path...)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the path provided %v doesn't exists on the object %s", path, template.GetName())
	}
	return nil
}

func (u *capiResourceUpdater) ApplyTemplate(ctx context.Context, template *unstructured.Unstructured, values map[string]interface{}, dryRun bool) error {
	_, err := u.UpdateTemplate(template, values)
	if err != nil {
		return err
	}
	u.Log.Info("Applying patch", "object", template.GetName(), "kind", template.GetKind(), "values", values)
	err = u.ApplyPatch(ctx, template, dryRun)
	if err != nil {
		return err
	}
	return nil
}

func (u *capiResourceUpdater) ApplyUpdatedTemplate(ctx context.Context, template *unstructured.Unstructured, dryRun bool) error {
	u.Log.Info("applied template", "object", template.GetName(), "kind", template.GetKind(), "dryRun", dryRun)
	dryRunStage := []string{}
	if dryRun {
		dryRunStage = []string{"All"}
	}
	err := u.client.Update(ctx, template, &client.UpdateOptions{FieldManager: "eks-controller", DryRun: dryRunStage})
	if err != nil {
		return err
	}
	return nil
}

func (u *capiResourceUpdater) ForceApplyTemplate(ctx context.Context, template *unstructured.Unstructured, dryRun bool) error {
	u.Log.Info("force applying template", "object", template.GetName(), "kind", template.GetKind(), "dryRun", dryRun)
	force := true
	dryRunStage := []string{}
	if dryRun {
		dryRunStage = []string{"All"}
	}
	err := u.client.Patch(ctx, template, client.Apply, &client.PatchOptions{FieldManager: "eks-controller", DryRun: dryRunStage, Force: &force})
	if err != nil {
		return err
	}
	return nil
}
