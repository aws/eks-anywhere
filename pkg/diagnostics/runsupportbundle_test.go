package diagnostics_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
	supportMocks "github.com/aws/eks-anywhere/pkg/diagnostics/interfaces/mocks"
	"github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers"
	providerMocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
)

func TestParseTimeOptions(t *testing.T) {
	type args struct {
		since     string
		sinceTime string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Without time options",
			args: args{
				since:     "",
				sinceTime: "",
			},
			wantErr: false,
		},
		{
			name: "Good since options",
			args: args{
				since:     "8h",
				sinceTime: "",
			},
			wantErr: false,
		},
		{
			name: "Good since time options",
			args: args{
				since:     "",
				sinceTime: "2021-06-28T15:04:05Z",
			},
			wantErr: false,
		},
		{
			name: "Duplicate time options",
			args: args{
				since:     "8m",
				sinceTime: "2021-06-28T15:04:05Z",
			},
			wantErr: true,
		},
		{
			name: "Wrong since time options",
			args: args{
				since:     "",
				sinceTime: "2021-06-28T15:04:05Z07:00",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := diagnostics.ParseTimeOptions(tt.args.since, tt.args.sinceTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimeOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGenerateBundleConfigWithExternalEtcd(t *testing.T) {
	spec := &cluster.Spec{
		Cluster: &eksav1alpha1.Cluster{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       eksav1alpha1.ClusterSpec{},
			Status:     eksav1alpha1.ClusterStatus{},
		},
	}

	spec.Cluster.Spec.ExternalEtcdConfiguration = &eksav1alpha1.ExternalEtcdConfiguration{
		Count: 3,
		MachineGroupRef: &eksav1alpha1.Ref{
			Kind: eksav1alpha1.VSphereMachineConfigKind,
			Name: "testRef",
		},
	}

	spec.Cluster.Spec.DatacenterRef = eksav1alpha1.Ref{
		Kind: eksav1alpha1.VSphereDatacenterKind,
		Name: "testRef",
	}

	t.Run(t.Name(), func(t *testing.T) {
		p := givenProvider(t)
		p.EXPECT().MachineConfigs().Return(machineConfigs())

		a := givenMockAnalyzerFactory(t)
		a.EXPECT().EksaExternalEtcdAnalyzers().Return(nil)
		a.EXPECT().DataCenterConfigAnalyzers(spec.Cluster.Spec.DatacenterRef).Return(nil)
		a.EXPECT().DefaultAnalyzers().Return(nil)

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)
		c.EXPECT().EksaHostCollectors(gomock.Any()).Return(nil)

		w := givenWriter(t)
		w.EXPECT().Write(gomock.Any(), gomock.Any())

		opts := diagnostics.EksaDiagnosticBundleOpts{
			AnalyzerFactory:  a,
			CollectorFactory: c,
			Writer:           w,
		}

		_, _ = diagnostics.NewDiagnosticBundleFromSpec(spec, p, "", opts)
	})
}

func TestGenerateBundleConfigWithOidc(t *testing.T) {
	spec := &cluster.Spec{
		Cluster: &eksav1alpha1.Cluster{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       eksav1alpha1.ClusterSpec{},
			Status:     eksav1alpha1.ClusterStatus{},
		},
	}

	spec.OIDCConfig = &eksav1alpha1.OIDCConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       eksav1alpha1.OIDCConfigSpec{},
		Status:     eksav1alpha1.OIDCConfigStatus{},
	}

	spec.Cluster.Spec.DatacenterRef = eksav1alpha1.Ref{
		Kind: eksav1alpha1.VSphereDatacenterKind,
		Name: "testRef",
	}

	t.Run(t.Name(), func(t *testing.T) {
		p := givenProvider(t)
		p.EXPECT().MachineConfigs().Return(machineConfigs())

		a := givenMockAnalyzerFactory(t)
		a.EXPECT().EksaOidcAnalyzers().Return(nil)
		a.EXPECT().DataCenterConfigAnalyzers(spec.Cluster.Spec.DatacenterRef).Return(nil)
		a.EXPECT().DefaultAnalyzers().Return(nil)

		w := givenWriter(t)
		w.EXPECT().Write(gomock.Any(), gomock.Any())

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)
		c.EXPECT().EksaHostCollectors(gomock.Any()).Return(nil)

		opts := diagnostics.EksaDiagnosticBundleOpts{
			AnalyzerFactory:  a,
			CollectorFactory: c,
			Writer:           w,
		}

		_, _ = diagnostics.NewDiagnosticBundleFromSpec(spec, p, "", opts)
	})
}

func TestGenerateBundleConfigWithGitOps(t *testing.T) {
	spec := &cluster.Spec{
		Cluster: &eksav1alpha1.Cluster{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       eksav1alpha1.ClusterSpec{},
			Status:     eksav1alpha1.ClusterStatus{},
		},
	}

	spec.GitOpsConfig = &eksav1alpha1.GitOpsConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       eksav1alpha1.GitOpsConfigSpec{},
		Status:     eksav1alpha1.GitOpsConfigStatus{},
	}

	spec.Cluster.Spec.DatacenterRef = eksav1alpha1.Ref{
		Kind: eksav1alpha1.VSphereDatacenterKind,
		Name: "testRef",
	}

	t.Run(t.Name(), func(t *testing.T) {
		p := givenProvider(t)
		p.EXPECT().MachineConfigs().Return(machineConfigs())

		a := givenMockAnalyzerFactory(t)
		a.EXPECT().EksaGitopsAnalyzers().Return(nil)
		a.EXPECT().DataCenterConfigAnalyzers(spec.Cluster.Spec.DatacenterRef).Return(nil)
		a.EXPECT().DefaultAnalyzers().Return(nil)

		w := givenWriter(t)
		w.EXPECT().Write(gomock.Any(), gomock.Any())

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)
		c.EXPECT().EksaHostCollectors(gomock.Any()).Return(nil)

		opts := diagnostics.EksaDiagnosticBundleOpts{
			AnalyzerFactory:  a,
			CollectorFactory: c,
			Writer:           w,
		}

		_, _ = diagnostics.NewDiagnosticBundleFromSpec(spec, p, "", opts)
	})
}

func TestGenerateDefaultBundle(t *testing.T) {
	t.Run(t.Name(), func(t *testing.T) {
		a := givenMockAnalyzerFactory(t)
		a.EXPECT().DefaultAnalyzers().Return(nil)

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)

		_ = diagnostics.NewDiagnosticBundleDefault(a, c)
	})
}

func TestGenerateCustomBundle(t *testing.T) {
	t.Run(t.Name(), func(t *testing.T) {
		_ = diagnostics.NewDiagnosticBundleCustom("", "", getOpts(t))
	})
}

func givenMockAnalyzerFactory(t *testing.T) *supportMocks.MockAnalyzerFactory {
	ctrl := gomock.NewController(t)
	return supportMocks.NewMockAnalyzerFactory(ctrl)
}

func givenMockCollectorsFactory(t *testing.T) *supportMocks.MockCollectorFactory {
	ctrl := gomock.NewController(t)
	return supportMocks.NewMockCollectorFactory(ctrl)
}

func getOpts(t *testing.T) diagnostics.EksaDiagnosticBundleOpts {
	return diagnostics.EksaDiagnosticBundleOpts{
		AnalyzerFactory:  givenMockAnalyzerFactory(t),
		CollectorFactory: givenMockCollectorsFactory(t),
	}
}

func givenWriter(t *testing.T) *mocks.MockFileWriter {
	ctrl := gomock.NewController(t)
	return mocks.NewMockFileWriter(ctrl)
}

func givenProvider(t *testing.T) *providerMocks.MockProvider {
	ctrl := gomock.NewController(t)
	return providerMocks.NewMockProvider(ctrl)
}

func machineConfigs() []providers.MachineConfig {
	var m []providers.MachineConfig
	return m
}
