package v1alpha1

import (
	"testing"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
)

func TestGetTinkerbellTemplateConfig(t *testing.T) {
	tests := []struct {
		testName    string
		fileName    string
		wantConfigs map[string]*TinkerbellTemplateConfig
		wantErr     bool
	}{
		{
			testName:    "file doesn't exist",
			fileName:    "testdata/fake_file.yaml",
			wantConfigs: nil,
			wantErr:     true,
		},
		{
			testName:    "not parseable file",
			fileName:    "testdata/not_parseable_cluster_tinkerbell.yaml",
			wantConfigs: nil,
			wantErr:     true,
		},
		{
			testName: "valid tinkerbell template config",
			fileName: "testdata/cluster_1_21_valid_tinkerbell.yaml",
			wantConfigs: map[string]*TinkerbellTemplateConfig{
				"tink-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       TinkerbellTemplateConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "tink-test",
					},
					Spec: TinkerbellTemplateConfigSpec{
						Template: tinkerbell.Workflow{
							Version:       "0.1",
							Name:          "tink-test",
							GlobalTimeout: 6000,
							ID:            "",
							Tasks: []tinkerbell.Task{
								{
									Name:       "tink-test",
									WorkerAddr: "{{.device_1}}",
									Volumes: []string{
										"/dev:/dev",
										"/dev/console:/dev/console",
										"/lib/firmware:/lib/firmware:ro",
									},
									Actions: []tinkerbell.Action{
										{
											Name:    "stream-image",
											Image:   "image2disk:v1.0.0",
											Timeout: 600,
											Environment: map[string]string{
												"IMG_URL":    "",
												"DEST_DISK":  "/dev/sda",
												"COMPRESSED": "true",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "multiple tinkerbell template configs",
			fileName: "testdata/cluster_1_21_valid_multiple_tinkerbell_templates.yaml",
			wantConfigs: map[string]*TinkerbellTemplateConfig{
				"tink-test-1": {
					TypeMeta: metav1.TypeMeta{
						Kind:       TinkerbellTemplateConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "tink-test-1",
					},
					Spec: TinkerbellTemplateConfigSpec{
						Template: tinkerbell.Workflow{
							Version:       "0.1",
							Name:          "tink-test-1",
							GlobalTimeout: 6000,
							ID:            "",
							Tasks: []tinkerbell.Task{
								{
									Name:       "tink-test-1",
									WorkerAddr: "{{.device_1}}",
									Volumes: []string{
										"/dev:/dev",
										"/dev/console:/dev/console",
										"/lib/firmware:/lib/firmware:ro",
									},
									Actions: []tinkerbell.Action{
										{
											Name:    "stream-image",
											Image:   "image2disk:v1.0.0",
											Timeout: 600,
											Environment: map[string]string{
												"IMG_URL":    "",
												"DEST_DISK":  "/dev/sda",
												"COMPRESSED": "true",
											},
										},
									},
								},
							},
						},
					},
				},
				"tink-test-2": {
					TypeMeta: metav1.TypeMeta{
						Kind:       TinkerbellTemplateConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "tink-test-2",
					},
					Spec: TinkerbellTemplateConfigSpec{
						Template: tinkerbell.Workflow{
							Version:       "0.1",
							Name:          "tink-test-2",
							GlobalTimeout: 6000,
							ID:            "",
							Tasks: []tinkerbell.Task{
								{
									Name:       "tink-test-2",
									WorkerAddr: "{{.device_1}}",
									Volumes: []string{
										"/dev:/dev",
										"/dev/console:/dev/console",
										"/lib/firmware:/lib/firmware:ro",
									},
									Actions: []tinkerbell.Action{
										{
											Name:    "stream-image",
											Image:   "image2disk:v1.0.0",
											Timeout: 600,
											Environment: map[string]string{
												"IMG_URL":    "",
												"DEST_DISK":  "/dev/sda",
												"COMPRESSED": "true",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			g := gomega.NewWithT(t)
			got, err := GetTinkerbellTemplateConfig(tt.fileName)
			g.Expect((err != nil)).To(gomega.BeEquivalentTo(tt.wantErr))
			g.Expect(got).To(gomega.BeEquivalentTo(tt.wantConfigs))
		})
	}
}
