package v1alpha1

import (
	"testing"
)

func TestUserConfigRefSameConfig(t *testing.T) {
	userConfigOld := []UserConfiguration{
		{
			Name: "testUser",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test",
			},
		},
		{
			Name: "testUser2",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test2",
			},
		},
	}
	userConfigNew := []UserConfiguration{
		{
			Name: "testUser",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test",
			},
		},
		{
			Name: "testUser2",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test2",
			},
		},
	}

	got := UsersSliceEqual(userConfigOld, userConfigNew)
	want := true
	if got != want {
		t.Fatalf("Expected %v, got %v", want, got)
	}
}

func TestUserConfigRefSameConfigSwapSsh(t *testing.T) {
	userConfigOld := []UserConfiguration{
		{
			Name: "testUser",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test",
				"ssh_rsa_test2",
			},
		},
		{
			Name: "testUser2",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test2",
			},
		},
	}
	userConfigNew := []UserConfiguration{
		{
			Name: "testUser",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test2",
				"ssh_rsa_test",
			},
		},
		{
			Name: "testUser2",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test2",
			},
		},
	}

	got := UsersSliceEqual(userConfigOld, userConfigNew)
	want := true
	if got != want {
		t.Fatalf("Expected %v, got %v", want, got)
	}
}

func TestUserConfigRefSameConfigSwapUser(t *testing.T) {
	userConfigOld := []UserConfiguration{
		{
			Name: "testUser2",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test2",
			},
		},
		{
			Name: "testUser",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test",
			},
		},
	}
	userConfigNew := []UserConfiguration{
		{
			Name: "testUser",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test",
			},
		},
		{
			Name: "testUser2",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test2",
			},
		},
	}

	got := UsersSliceEqual(userConfigOld, userConfigNew)
	want := true
	if got != want {
		t.Fatalf("Expected %v, got %v", want, got)
	}
}

func TestUserConfigOldRefSDiffConfig(t *testing.T) {
	userConfigOld := []UserConfiguration{
		{
			Name: "testUser",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test",
			},
		},
		{
			Name: "testUser2",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test2",
			},
		},
	}
	userConfigNew := []UserConfiguration{
		{
			Name: "testUserChanged",
			SshAuthorizedKeys: []string{
				"ssh_rsa_testChanged",
			},
		},
		{
			Name: "testUser2",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test2",
			},
		},
	}

	got := UsersSliceEqual(userConfigOld, userConfigNew)
	want := false
	if got != want {
		t.Fatalf("Expected %v, got %v", want, got)
	}
}

func TestUserConfigNewRefEmptyConfig(t *testing.T) {
	userConfigOld := []UserConfiguration{
		{
			Name: "testUser",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test",
			},
		},
	}
	var userConfigNew []UserConfiguration

	got := UsersSliceEqual(userConfigOld, userConfigNew)
	want := false
	if got != want {
		t.Fatalf("Expected %v, got %v", want, got)
	}
}

func TestUserConfigRefEmptyConfig(t *testing.T) {
	var userConfigOld []UserConfiguration
	userConfigNew := []UserConfiguration{
		{
			Name: "testUser",
			SshAuthorizedKeys: []string{
				"ssh_rsa_test",
			},
		},
	}

	got := UsersSliceEqual(userConfigOld, userConfigNew)
	want := false
	if got != want {
		t.Fatalf("Expected %v, got %v", want, got)
	}
}
