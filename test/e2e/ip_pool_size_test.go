package e2e

import (
	"regexp"
	"testing"
)

// makeReq creates a single IPPoolRequirement for testing.
func makeReq(pattern string, size int) IPPoolRequirement {
	return IPPoolRequirement{
		re:         regexp.MustCompile(pattern),
		ipPoolSize: size,
	}
}

func TestGetIPPoolSizeFirstMatchWins(t *testing.T) {
	reqs := []IPPoolRequirement{
		makeReq("^TestFooSpecific$", 5),
		makeReq("Foo", 2),
	}

	// Specific pattern should match first
	size := GetIPPoolSize("TestFooSpecific", reqs)
	if size != 5 {
		t.Errorf("GetIPPoolSize(TestFooSpecific) = %d, want 5", size)
	}

	// Generic pattern should match when specific doesn't
	size = GetIPPoolSize("TestFooGeneric", reqs)
	if size != 2 {
		t.Errorf("GetIPPoolSize(TestFooGeneric) = %d, want 2", size)
	}
}

func TestGetIPPoolSizeDefaultForNoMatch(t *testing.T) {
	reqs := []IPPoolRequirement{
		makeReq("Foo", 3),
	}

	size := GetIPPoolSize("TestBarUnrelated", reqs)
	if size != defaultIPPoolSize {
		t.Errorf("GetIPPoolSize(TestBarUnrelated) = %d, want %d (default)", size, defaultIPPoolSize)
	}
}

func TestGetIPPoolSizeEmptyRequirements(t *testing.T) {
	size := GetIPPoolSize("TestAnything", nil)
	if size != defaultIPPoolSize {
		t.Errorf("GetIPPoolSize with nil requirements = %d, want %d (default)", size, defaultIPPoolSize)
	}
}

func TestGetIPPoolSizeMultiplePatterns(t *testing.T) {
	reqs := []IPPoolRequirement{
		makeReq(`^TestA\d+Big$`, 5),
		makeReq(`^TestA\d+Medium$`, 3),
		makeReq(`^TestA`, 2),
	}

	tests := []struct {
		testName string
		want     int
	}{
		{"TestA123Big", 5},
		{"TestA456Medium", 3},
		{"TestAOther", 2},
		{"TestBSomething", 1},
	}
	for _, tt := range tests {
		got := GetIPPoolSize(tt.testName, reqs)
		if got != tt.want {
			t.Errorf("GetIPPoolSize(%q) = %d, want %d", tt.testName, got, tt.want)
		}
	}
}

// TestLoadIPPoolRequirements verifies the embedded YAML can be loaded without errors.
func TestLoadIPPoolRequirements(t *testing.T) {
	reqs, err := LoadIPPoolRequirements()
	if err != nil {
		t.Fatalf("LoadIPPoolRequirements() error = %v", err)
	}
	if len(reqs) == 0 {
		t.Fatal("LoadIPPoolRequirements() returned empty requirements")
	}
}
