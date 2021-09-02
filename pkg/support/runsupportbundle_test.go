package supportbundle_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	support "github.com/aws/eks-anywhere/pkg/support"
	"github.com/aws/eks-anywhere/pkg/support/interfaces/mocks"
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
			_, err := support.ParseTimeOptions(tt.args.since, tt.args.sinceTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimeOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestParseBundleFromDoc(t *testing.T) {
	type args struct {
		bundleConfig string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Good bundle config",
			args: args{
				bundleConfig: "testdata/support-bundle-test1.yaml",
			},
			wantErr: false,
		},
		{
			name: "Wrong bundle config",
			args: args{
				bundleConfig: "testdata/support-bundle-test2.yaml",
			},
			wantErr: true,
		},
	}

	spec := &cluster.Spec{
		Cluster: &eksav1alpha1.Cluster{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       eksav1alpha1.ClusterSpec{},
			Status:     eksav1alpha1.ClusterStatus{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := support.ParseBundleFromDoc(spec, tt.args.bundleConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBundleFromDoc() error = %v, wantErr %v", err, tt.wantErr)
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
		a := givenMockAnalyzerFactory(t)
		a.EXPECT().EksaExternalEtcdAnalyzers().Return(nil)
		a.EXPECT().DataCenterConfigAnalyzers(spec.Cluster.Spec.DatacenterRef).Return(nil)
		a.EXPECT().DefaultAnalyzers().Return(nil)

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)

		_ = support.NewBundleConfig(spec, a, c)
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
		a := givenMockAnalyzerFactory(t)
		a.EXPECT().EksaOidcAnalyzers().Return(nil)
		a.EXPECT().DataCenterConfigAnalyzers(spec.Cluster.Spec.DatacenterRef).Return(nil)
		a.EXPECT().DefaultAnalyzers().Return(nil)

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)

		_ = support.NewBundleConfig(spec, a, c)
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
		a := givenMockAnalyzerFactory(t)
		a.EXPECT().EksaGitopsAnalyzers().Return(nil)
		a.EXPECT().DataCenterConfigAnalyzers(spec.Cluster.Spec.DatacenterRef).Return(nil)
		a.EXPECT().DefaultAnalyzers().Return(nil)

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)

		_ = support.NewBundleConfig(spec, a, c)
	})
}

func TestGenerateDefaultBundle(t *testing.T) {
	t.Run(t.Name(), func(t *testing.T) {
		a := givenMockAnalyzerFactory(t)
		a.EXPECT().DefaultAnalyzers().Return(nil)

		c := givenMockCollectorsFactory(t)
		c.EXPECT().DefaultCollectors().Return(nil)

		_ = support.NewDefaultBundleConfig(a, c)
	})
}

func TestGenerateCustomBundle(t *testing.T) {
	bundle := &v1beta2.SupportBundle{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1beta2.SupportBundleSpec{},
		Status:     v1beta2.SupportBundleStatus{},
	}

	t.Run(t.Name(), func(t *testing.T) {
		a := givenMockAnalyzerFactory(t)
		c := givenMockCollectorsFactory(t)

		_ = support.NewCustomBundleConfig(bundle, a, c)
	})
}

func givenMockAnalyzerFactory(t *testing.T) *mocks.MockAnalyzerFactory {
	ctrl := gomock.NewController(t)
	return mocks.NewMockAnalyzerFactory(ctrl)
}

func givenMockCollectorsFactory(t *testing.T) *mocks.MockCollectorFactory {
	ctrl := gomock.NewController(t)
	return mocks.NewMockCollectorFactory(ctrl)
}
