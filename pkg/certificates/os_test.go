package certificates_test

import (
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/pkg/certificates"
)

func TestBuildOSRenewer_Linux(t *testing.T) {
	got := certificates.BuildOSRenewer(string(certificates.OSTypeLinux), t.TempDir())

	wantType := reflect.TypeOf(&certificates.LinuxRenewer{})
	gotType := reflect.TypeOf(got)
	if gotType != wantType {
		t.Fatalf("BuildOSRenewer() for linux expected type %v, got: %v", wantType, gotType)
	}
}

func TestBuildOSRenewer_Bottlerocket(t *testing.T) {
	got := certificates.BuildOSRenewer(string(certificates.OSTypeBottlerocket), t.TempDir())

	wantType := reflect.TypeOf(&certificates.BottlerocketRenewer{})
	gotType := reflect.TypeOf(got)
	if gotType != wantType {
		t.Fatalf("BuildOSRenewer() for bottlerocket expected type %v, got: %v", wantType, gotType)
	}
}

func TestBuildOSRenewer_UnknownOSPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("BuildOSRenewer() expected panic for unknown OS type, got none")
		}
	}()

	certificates.BuildOSRenewer("unknown", t.TempDir())
}
