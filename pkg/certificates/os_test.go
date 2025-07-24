package certificates

import (
	"reflect"
	"testing"
)

func TestBuildOSRenewer_Linux(t *testing.T) {
	got := BuildOSRenewer(string(OSTypeLinux), t.TempDir())

	wantType := reflect.TypeOf(&LinuxRenewer{})
	gotType := reflect.TypeOf(got)
	if gotType != wantType {
		t.Fatalf("BuildOSRenewer() for linux expected type %v, got: %v", wantType, gotType)
	}
}

func TestBuildOSRenewer_Bottlerocket(t *testing.T) {
	got := BuildOSRenewer(string(OSTypeBottlerocket), t.TempDir())

	wantType := reflect.TypeOf(&BottlerocketRenewer{})
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

	BuildOSRenewer("unknown", t.TempDir())
}
