package resource

import (
	"testing"

	. "github.com/onsi/gomega"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	key1WithComment = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8ZEibIrz1AUBKDvmDiWLs9f5DnOerC4qPITiDtSOuPAsxgZbRMavBfVTxodMdAkYRYlXxK6PqNo0ve0qcOV2yvpxH1OogasMMetck6BlM/dIoo3vEY4ZoG9DuVRIf9Iry5gJKbpMDYWpx1IGZrDMOFcIM20ii2qLQQk5hfq9OqdqhToEJFixdgJt/y/zt6Koy3kix+XsnrVdAHgWAq4CZuwt1G6JUAqrpob3H8vPmL7aS+35ktf0pHBm6nYoxRhslnWMUb/7vpzWiq+fUBIm2LYqvrnm7t3fRqFx7p2sZqAm2jDNivyYXwRXkoQPR96zvGeMtuQ5BVGPpsDfVudSW21+pEXHI0GINtTbua7Ogz7wtpVywSvHraRgdFOeY9mkXPzvm2IhoqNrteck2GErwqSqb19mPz6LnHueK0u7i6WuQWJn0CUoCtyMGIrowXSviK8qgHXKrmfTWATmCkbtosnLskNdYuOw8bKxq5S4WgdQVhPps2TiMSZndjX5NTr8= ubuntu@ip-10-2-0-6"
	key1NoComment   = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8ZEibIrz1AUBKDvmDiWLs9f5DnOerC4qPITiDtSOuPAsxgZbRMavBfVTxodMdAkYRYlXxK6PqNo0ve0qcOV2yvpxH1OogasMMetck6BlM/dIoo3vEY4ZoG9DuVRIf9Iry5gJKbpMDYWpx1IGZrDMOFcIM20ii2qLQQk5hfq9OqdqhToEJFixdgJt/y/zt6Koy3kix+XsnrVdAHgWAq4CZuwt1G6JUAqrpob3H8vPmL7aS+35ktf0pHBm6nYoxRhslnWMUb/7vpzWiq+fUBIm2LYqvrnm7t3fRqFx7p2sZqAm2jDNivyYXwRXkoQPR96zvGeMtuQ5BVGPpsDfVudSW21+pEXHI0GINtTbua7Ogz7wtpVywSvHraRgdFOeY9mkXPzvm2IhoqNrteck2GErwqSqb19mPz6LnHueK0u7i6WuQWJn0CUoCtyMGIrowXSviK8qgHXKrmfTWATmCkbtosnLskNdYuOw8bKxq5S4WgdQVhPps2TiMSZndjX5NTr8="
	key2WithComment = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDACSIcJsmvL4KQ42+AFOwB4UoqwBTGZeSO2kol/ncmw/4OU07zJNuc+0pF7tk1G/9MbrrJsCK1uKsBIFKSwv+w4+LRBOaVKtjfVzzqKWQzYKRGlM2PRxDbovbHbcVQ4fYcIn3LwfuujZmYb2JX8/lKyL6ga/ZQWP/TNJF6M+y69mWElrhfAGKMYZrxhbuZUjGRp4a2fkrDxY3w8qnFP2Glyd5427j0WGt46G8zW8s6EP+ROc+0NCFqzqkFIJPONd5tUvEiwUtqS0FfgtO5Grv5VYh134FMUWqFG2/Ot/G1fwrMV+1UivS4iLDSGlDplHJHLg2OIsNlL6F2wHGb2N1ykVZAZ2tlhHM0oxdyNd5SEs4htvayEx9/u8RxRYcP2Mtgot401kCGyhq1CVzVsZBCOVw45+jrVZ7xuLgkxaiKBQycj3IsUjWpmkGL66F6HH8wtMN1YriZn2QRV6+z6eg7eI+yspSILL4P/Kw7/1Vkb2IuWUU5h9wE0Iac1yIF1pM= ubuntu@ip-10-2-0-8"
	key2NoComment   = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDACSIcJsmvL4KQ42+AFOwB4UoqwBTGZeSO2kol/ncmw/4OU07zJNuc+0pF7tk1G/9MbrrJsCK1uKsBIFKSwv+w4+LRBOaVKtjfVzzqKWQzYKRGlM2PRxDbovbHbcVQ4fYcIn3LwfuujZmYb2JX8/lKyL6ga/ZQWP/TNJF6M+y69mWElrhfAGKMYZrxhbuZUjGRp4a2fkrDxY3w8qnFP2Glyd5427j0WGt46G8zW8s6EP+ROc+0NCFqzqkFIJPONd5tUvEiwUtqS0FfgtO5Grv5VYh134FMUWqFG2/Ot/G1fwrMV+1UivS4iLDSGlDplHJHLg2OIsNlL6F2wHGb2N1ykVZAZ2tlhHM0oxdyNd5SEs4htvayEx9/u8RxRYcP2Mtgot401kCGyhq1CVzVsZBCOVw45+jrVZ7xuLgkxaiKBQycj3IsUjWpmkGL66F6HH8wtMN1YriZn2QRV6+z6eg7eI+yspSILL4P/Kw7/1Vkb2IuWUU5h9wE0Iac1yIF1pM="
)

func TestNeedsVSphereNewKubeadmConfigTemplate(t *testing.T) {
	tests := []struct {
		name               string
		newWorkerNodeGroup *anywherev1.WorkerNodeGroupConfiguration
		oldWorkerNodeGroup *anywherev1.WorkerNodeGroupConfiguration
		oldWorkerNodeVmc   *anywherev1.VSphereMachineConfig
		newWorkerNodeVmc   *anywherev1.VSphereMachineConfig
		want               bool
	}{
		{
			name:               "doesn't need new kubeadm",
			newWorkerNodeGroup: &anywherev1.WorkerNodeGroupConfiguration{},
			oldWorkerNodeGroup: &anywherev1.WorkerNodeGroupConfiguration{},
			newWorkerNodeVmc: &anywherev1.VSphereMachineConfig{
				Spec: anywherev1.VSphereMachineConfigSpec{
					Users: []anywherev1.UserConfiguration{
						{
							Name:              "user1",
							SshAuthorizedKeys: []string{key1WithComment, key2WithComment},
						},
					},
				},
			},
			oldWorkerNodeVmc: &anywherev1.VSphereMachineConfig{
				Spec: anywherev1.VSphereMachineConfigSpec{
					Users: []anywherev1.UserConfiguration{
						{
							Name:              "user1",
							SshAuthorizedKeys: []string{key1NoComment, key2NoComment},
						},
					},
				},
			},
			want: false,
		},
		{
			name:               "needs new kubeadm, different SSH keys",
			newWorkerNodeGroup: &anywherev1.WorkerNodeGroupConfiguration{},
			oldWorkerNodeGroup: &anywherev1.WorkerNodeGroupConfiguration{},
			newWorkerNodeVmc: &anywherev1.VSphereMachineConfig{
				Spec: anywherev1.VSphereMachineConfigSpec{
					Users: []anywherev1.UserConfiguration{
						{
							Name:              "user1",
							SshAuthorizedKeys: []string{"key1", "key2"},
						},
					},
				},
			},
			oldWorkerNodeVmc: &anywherev1.VSphereMachineConfig{
				Spec: anywherev1.VSphereMachineConfigSpec{
					Users: []anywherev1.UserConfiguration{
						{
							Name:              "user1",
							SshAuthorizedKeys: []string{"key1", "key2"},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(
				needsVSphereNewKubeadmConfigTemplate(
					tt.newWorkerNodeGroup,
					tt.oldWorkerNodeGroup,
					tt.oldWorkerNodeVmc,
					tt.newWorkerNodeVmc,
				),
			).To(Equal(tt.want))
		})
	}
}

func TestEquivalentUsers(t *testing.T) {
	tests := []struct {
		name string
		a    []anywherev1.UserConfiguration
		b    []anywherev1.UserConfiguration
		want bool
	}{
		{
			name: "identic",
			a: []anywherev1.UserConfiguration{
				{
					Name:              "user1",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
				{
					Name:              "user2",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
			},
			b: []anywherev1.UserConfiguration{
				{
					Name:              "user1",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
				{
					Name:              "user2",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
			},
			want: true,
		},
		{
			name: "same size, different users",
			a: []anywherev1.UserConfiguration{
				{
					Name:              "user1",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
				{
					Name:              "user2",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
			},
			b: []anywherev1.UserConfiguration{
				{
					Name:              "user1",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
				{
					Name:              "user3",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
			},
			want: false,
		},
		{
			name: "same size, different keys",
			a: []anywherev1.UserConfiguration{
				{
					Name:              "user1",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
				{
					Name:              "user2",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
			},
			b: []anywherev1.UserConfiguration{
				{
					Name:              "user1",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
				{
					Name:              "user2",
					SshAuthorizedKeys: []string{"key1", "key3"},
				},
			},
			want: false,
		},
		{
			name: "different sizes",
			a: []anywherev1.UserConfiguration{
				{
					Name:              "user1",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
				{
					Name:              "user2",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
			},
			b: []anywherev1.UserConfiguration{
				{
					Name:              "user1",
					SshAuthorizedKeys: []string{"key1", "key2"},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(equivalentUsers(tt.a, tt.b)).To(Equal(tt.want))
		})
	}
}

func TestEquivalentSSHKeys(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "identic",
			a:    []string{key1WithComment, key2NoComment},
			b:    []string{key1WithComment, key2NoComment},
			want: true,
		},
		{
			name: "equal, different order",
			a:    []string{key1WithComment, key2NoComment},
			b:    []string{key2NoComment, key1WithComment},
			want: true,
		},
		{
			name: "equal, invalid key",
			a:    []string{"key"},
			b:    []string{"key"},
			want: true,
		},
		{
			name: "same size, different keys",
			a:    []string{"key", "key2"},
			b:    []string{"key", "key3"},
			want: false,
		},
		{
			name: "different sizes",
			a:    []string{"key"},
			b:    []string{"key", "key3"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(equivalentSSHKeys(tt.a, tt.b)).To(Equal(tt.want))
		})
	}
}
