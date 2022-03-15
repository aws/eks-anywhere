package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetSnowMachineConfigs(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     map[string]*SnowMachineConfig
		wantErr  string
	}{
		{
			name:     "file doesn't exist",
			fileName: "testdata/fake_file.yaml",
			want:     nil,
			wantErr:  "no such file or directory",
		},
		{
			name:     "not parseable file",
			fileName: "testdata/not_parseable_cluster_snow.yaml",
			want:     nil,
			wantErr:  "error unmarshaling JSON: while decoding JSON: json: unknown field",
		},
		{
			name:     "valid 1.21",
			fileName: "testdata/cluster_1_21_snow.yaml",
			want: map[string]*SnowMachineConfig{
				"eksa-unit-test": {
					TypeMeta: metav1.TypeMeta{
						Kind:       SnowMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: SnowMachineConfigSpec{
						AMIID:        "eks-d-v1-21-5-ubuntu-ami-02833ca9a8f29c2ea",
						InstanceType: "sbe-c.large",
						SshKeyName:   "default",
					},
				},
			},
			wantErr: "",
		},
		{
			name:     "valid 1.21 with multiple machine configs",
			fileName: "testdata/cluster_1_21_snow_multiple_machineconfigs.yaml",
			want: map[string]*SnowMachineConfig{
				"eksa-unit-test-cp": {
					TypeMeta: metav1.TypeMeta{
						Kind:       SnowMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test-cp",
					},
					Spec: SnowMachineConfigSpec{
						AMIID:        "eks-d-v1-21-5-ubuntu-ami-02833ca9a8f29c2ea",
						InstanceType: "sbe-c.large",
						SshKeyName:   "default",
					},
				},
				"eksa-unit-test-md": {
					TypeMeta: metav1.TypeMeta{
						Kind:       SnowMachineConfigKind,
						APIVersion: SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test-md",
					},
					Spec: SnowMachineConfigSpec{
						AMIID:        "eks-d-v1-21-5-ubuntu-ami-02833ca9a8f29c2ea",
						InstanceType: "sbe-c.xlarge",
						SshKeyName:   "default",
					},
				},
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got, err := GetSnowMachineConfigs(tt.fileName)
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}

			g.Expect(got).To(Equal(tt.want))
		})
	}
}
