package diagnostics_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
	supportMocks "github.com/aws/eks-anywhere/pkg/diagnostics/interfaces/mocks"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
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
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &eksav1alpha1.Cluster{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec: eksav1alpha1.ClusterSpec{
				DatacenterRef: eksav1alpha1.Ref{
					Kind: eksav1alpha1.VSphereDatacenterKind,
					Name: "testRef",
				},
				ExternalEtcdConfiguration: &eksav1alpha1.ExternalEtcdConfiguration{
					Count: 3,
					MachineGroupRef: &eksav1alpha1.Ref{
						Kind: eksav1alpha1.VSphereMachineConfigKind,
						Name: "testRef",
					},
				},
			},
			Status: eksav1alpha1.ClusterStatus{},
		}
	})

	t.Run(t.Name(), func(t *testing.T) {
		p := givenProvider(t)
		p.EXPECT().MachineConfigs(spec).Return(machineConfigs())

		a := givenMockAnalyzerFactory(t)
		a.EXPECT().EksaExternalEtcdAnalyzers().Return(nil)
		a.EXPECT().DataCenterConfigAnalyzers(spec.Cluster.Spec.DatacenterRef).Return(nil)
		a.EXPECT().DefaultAnalyzers().Return(nil)
		a.EXPECT().EksaLogTextAnalyzers(gomock.Any()).Return(nil)
		a.EXPECT().ManagementClusterAnalyzers().Return(nil)

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)
		c.EXPECT().EksaHostCollectors(gomock.Any()).Return(nil)
		c.EXPECT().ManagementClusterCollectors().Return(nil)
		c.EXPECT().DataCenterConfigCollectors(spec.Cluster.Spec.DatacenterRef).Return(nil)

		w := givenWriter(t)
		w.EXPECT().Write(gomock.Any(), gomock.Any())

		opts := diagnostics.EksaDiagnosticBundleFactoryOpts{
			AnalyzerFactory:  a,
			CollectorFactory: c,
			Writer:           w,
		}

		f := diagnostics.NewFactory(opts)
		_, _ = f.DiagnosticBundleFromSpec(spec, p, "")
	})
}

func TestGenerateBundleConfigWithOidc(t *testing.T) {
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &eksav1alpha1.Cluster{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec: eksav1alpha1.ClusterSpec{
				DatacenterRef: eksav1alpha1.Ref{
					Kind: eksav1alpha1.VSphereDatacenterKind,
					Name: "testRef",
				},
			},
			Status: eksav1alpha1.ClusterStatus{},
		}
		s.OIDCConfig = &eksav1alpha1.OIDCConfig{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       eksav1alpha1.OIDCConfigSpec{},
			Status:     eksav1alpha1.OIDCConfigStatus{},
		}
	})

	t.Run(t.Name(), func(t *testing.T) {
		p := givenProvider(t)
		p.EXPECT().MachineConfigs(spec).Return(machineConfigs())

		a := givenMockAnalyzerFactory(t)
		a.EXPECT().EksaOidcAnalyzers().Return(nil)
		a.EXPECT().DataCenterConfigAnalyzers(spec.Cluster.Spec.DatacenterRef).Return(nil)
		a.EXPECT().DefaultAnalyzers().Return(nil)
		a.EXPECT().EksaLogTextAnalyzers(gomock.Any()).Return(nil)
		a.EXPECT().ManagementClusterAnalyzers().Return(nil)

		w := givenWriter(t)
		w.EXPECT().Write(gomock.Any(), gomock.Any())

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)
		c.EXPECT().EksaHostCollectors(gomock.Any()).Return(nil)
		c.EXPECT().ManagementClusterCollectors().Return(nil)
		c.EXPECT().DataCenterConfigCollectors(spec.Cluster.Spec.DatacenterRef).Return(nil)

		opts := diagnostics.EksaDiagnosticBundleFactoryOpts{
			AnalyzerFactory:  a,
			CollectorFactory: c,
			Writer:           w,
		}

		f := diagnostics.NewFactory(opts)
		_, _ = f.DiagnosticBundleFromSpec(spec, p, "")
	})
}

func TestGenerateBundleConfigWithGitOps(t *testing.T) {
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &eksav1alpha1.Cluster{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec: eksav1alpha1.ClusterSpec{
				DatacenterRef: eksav1alpha1.Ref{
					Kind: eksav1alpha1.VSphereDatacenterKind,
					Name: "testRef",
				},
			},
			Status: eksav1alpha1.ClusterStatus{},
		}
		s.GitOpsConfig = &eksav1alpha1.GitOpsConfig{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       eksav1alpha1.GitOpsConfigSpec{},
			Status:     eksav1alpha1.GitOpsConfigStatus{},
		}
	})

	t.Run(t.Name(), func(t *testing.T) {
		p := givenProvider(t)
		p.EXPECT().MachineConfigs(spec).Return(machineConfigs())

		a := givenMockAnalyzerFactory(t)
		a.EXPECT().EksaGitopsAnalyzers().Return(nil)
		a.EXPECT().DataCenterConfigAnalyzers(spec.Cluster.Spec.DatacenterRef).Return(nil)
		a.EXPECT().DefaultAnalyzers().Return(nil)
		a.EXPECT().EksaLogTextAnalyzers(gomock.Any()).Return(nil)
		a.EXPECT().ManagementClusterAnalyzers().Return(nil)

		w := givenWriter(t)
		w.EXPECT().Write(gomock.Any(), gomock.Any())

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)
		c.EXPECT().EksaHostCollectors(gomock.Any()).Return(nil)
		c.EXPECT().ManagementClusterCollectors().Return(nil)
		c.EXPECT().DataCenterConfigCollectors(spec.Cluster.Spec.DatacenterRef).Return(nil)

		opts := diagnostics.EksaDiagnosticBundleFactoryOpts{
			AnalyzerFactory:  a,
			CollectorFactory: c,
			Writer:           w,
		}

		f := diagnostics.NewFactory(opts)
		_, _ = f.DiagnosticBundleFromSpec(spec, p, "")
	})
}

func TestGenerateDefaultBundle(t *testing.T) {
	t.Run(t.Name(), func(t *testing.T) {
		a := givenMockAnalyzerFactory(t)
		a.EXPECT().DefaultAnalyzers().Return(nil)
		a.EXPECT().ManagementClusterAnalyzers().Return(nil)

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)
		c.EXPECT().ManagementClusterCollectors().Return(nil)

		w := givenWriter(t)

		opts := diagnostics.EksaDiagnosticBundleFactoryOpts{
			AnalyzerFactory:  a,
			CollectorFactory: c,
			Writer:           w,
		}

		f := diagnostics.NewFactory(opts)
		_ = f.DiagnosticBundleDefault()
	})
}

func TestBundleFromSpecComplete(t *testing.T) {
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &eksav1alpha1.Cluster{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec: eksav1alpha1.ClusterSpec{
				DatacenterRef: eksav1alpha1.Ref{
					Kind: eksav1alpha1.VSphereDatacenterKind,
					Name: "testRef",
				},
				ExternalEtcdConfiguration: &eksav1alpha1.ExternalEtcdConfiguration{
					Count: 3,
					MachineGroupRef: &eksav1alpha1.Ref{
						Kind: eksav1alpha1.VSphereMachineConfigKind,
						Name: "testRef",
					},
				},
			},
			Status: eksav1alpha1.ClusterStatus{},
		}
	})

	t.Run(t.Name(), func(t *testing.T) {
		ctx := context.Background()
		kubeconfig := "testcluster.kubeconfig"

		p := givenProvider(t)
		p.EXPECT().MachineConfigs(spec).Return(machineConfigs())

		a := givenMockAnalyzerFactory(t)
		a.EXPECT().EksaExternalEtcdAnalyzers().Return(nil)
		a.EXPECT().DataCenterConfigAnalyzers(spec.Cluster.Spec.DatacenterRef).Return(nil)
		a.EXPECT().DefaultAnalyzers().Return(nil)
		a.EXPECT().EksaLogTextAnalyzers(gomock.Any()).Return(nil)
		a.EXPECT().ManagementClusterAnalyzers().Return(nil)

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)
		c.EXPECT().EksaHostCollectors(gomock.Any()).Return(nil)
		c.EXPECT().ManagementClusterCollectors().Return(nil)
		c.EXPECT().DataCenterConfigCollectors(spec.Cluster.Spec.DatacenterRef).Return(nil)

		w := givenWriter(t)
		w.EXPECT().Write(gomock.Any(), gomock.Any()).Times(2)

		k, e := givenKubectl(t)
		expectedParam := []string{"create", "namespace", constants.EksaDiagnosticsNamespace, "--kubeconfig", kubeconfig}
		e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)

		expectedParam = []string{"delete", "namespace", constants.EksaDiagnosticsNamespace, "--kubeconfig", kubeconfig}
		e.EXPECT().Execute(ctx, gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)

		expectedParam = []string{"apply", "-f", "-", "--kubeconfig", kubeconfig}
		e.EXPECT().ExecuteWithStdin(ctx, gomock.Any(), gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)

		expectedParam = []string{"delete", "-f", "-", "--kubeconfig", kubeconfig}
		e.EXPECT().ExecuteWithStdin(ctx, gomock.Any(), gomock.Eq(expectedParam)).Return(bytes.Buffer{}, nil)

		returnAnalysis := []*executables.SupportBundleAnalysis{
			{
				Title:   "itsATestYo",
				IsPass:  true,
				IsFail:  false,
				IsWarn:  false,
				Message: "",
				Uri:     "",
			},
		}

		tc := givenTroubleshootClient(t)
		mockArchivePath := "/tmp/archive/path"
		tc.EXPECT().Collect(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(mockArchivePath, nil)
		tc.EXPECT().Analyze(ctx, gomock.Any(), mockArchivePath).Return(returnAnalysis, nil)

		opts := diagnostics.EksaDiagnosticBundleFactoryOpts{
			AnalyzerFactory:  a,
			CollectorFactory: c,
			Writer:           w,
			Kubectl:          k,
			Client:           tc,
		}

		var sinceTimeValue *time.Time
		sinceTimeValue, err := diagnostics.ParseTimeOptions("1h", "")
		if err != nil {
			t.Errorf("ParseTimeOptions() error = %v, wantErr nil", err)
			return
		}

		f := diagnostics.NewFactory(opts)
		b, _ := f.DiagnosticBundleFromSpec(spec, p, kubeconfig)
		err = b.CollectAndAnalyze(ctx, sinceTimeValue)
		if err != nil {
			t.Errorf("CollectAndAnalyze() error = %v, wantErr nil", err)
			return
		}
	})
}

func TestGenerateCustomBundle(t *testing.T) {
	t.Run(t.Name(), func(t *testing.T) {
		f := diagnostics.NewFactory(getOpts(t))
		_ = f.DiagnosticBundleCustom("", "")
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

func getOpts(t *testing.T) diagnostics.EksaDiagnosticBundleFactoryOpts {
	return diagnostics.EksaDiagnosticBundleFactoryOpts{
		AnalyzerFactory:  givenMockAnalyzerFactory(t),
		CollectorFactory: givenMockCollectorsFactory(t),
	}
}

func givenKubectl(t *testing.T) (*executables.Kubectl, *mockexecutables.MockExecutable) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)

	return executables.NewKubectl(executable), executable
}

func givenTroubleshootClient(t *testing.T) *supportMocks.MockBundleClient {
	ctrl := gomock.NewController(t)
	return supportMocks.NewMockBundleClient(ctrl)
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
