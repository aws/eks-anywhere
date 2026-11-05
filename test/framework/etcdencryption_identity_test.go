package framework

import (
	"bytes"
	"io"
	"path"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/aws/eks-anywhere/pkg/templater"
)

func TestKMSProviderIdentityWithoutWebhook(t *testing.T) {
	for _, controlPlaneOnly := range []string{"", "true"} {
		t.Run("controlPlaneOnly="+controlPlaneOnly, func(t *testing.T) {
			g := NewWithT(t)
			values := map[string]string{
				"kmsImage":           "example.invalid/encryption-provider@sha256:" + strings.Repeat("0", 64),
				"kmsIamRole":         "arn:aws:iam::123456789012:role/example-kms",
				"kmsKeyArn":          "arn:aws:kms:us-west-2:123456789012:key/example",
				"kmsKeyRegion":       "us-west-2",
				"kmsSocket":          "/var/run/kmsplugin/socket.sock",
				"serviceAccountName": "kms-encrypter-decrypter",
			}
			if controlPlaneOnly != "" {
				values["deployOnlyOnControlPlane"] = controlPlaneOnly
			}
			manifest, err := templater.Execute(kmsProviderManifest, values)
			g.Expect(err).NotTo(HaveOccurred())

			decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 4096)
			var serviceAccount corev1.ServiceAccount
			g.Expect(decoder.Decode(&serviceAccount)).To(Succeed())
			var daemonSet appsv1.DaemonSet
			g.Expect(decoder.Decode(&daemonSet)).To(Succeed())
			var extra interface{}
			g.Expect(decoder.Decode(&extra)).To(MatchError(io.EOF))

			pod := daemonSet.Spec.Template.Spec
			g.Expect(pod.ServiceAccountName).To(Equal(serviceAccount.Name))
			g.Expect(pod.Containers).To(HaveLen(1))
			container := pod.Containers[0]
			env := map[string]string{}
			for _, variable := range container.Env {
				env[variable.Name] = variable.Value
			}
			g.Expect(env["AWS_ROLE_ARN"]).To(Equal(serviceAccount.Annotations["eks.amazonaws.com/role-arn"]))
			g.Expect(env["AWS_ROLE_ARN"]).To(Equal(values["kmsIamRole"]))
			g.Expect(env["AWS_WEB_IDENTITY_TOKEN_FILE"]).NotTo(BeEmpty())
			g.Expect(env["AWS_STS_REGIONAL_ENDPOINTS"]).To(Equal("regional"))

			validateKMSProviderIdentityToken(g, pod, container, serviceAccount.Annotations["eks.amazonaws.com/audience"], env["AWS_WEB_IDENTITY_TOKEN_FILE"])
			g.Expect(pod.HostNetwork).To(BeFalse())
			g.Expect(pod.HostPID).To(BeFalse())
			for _, port := range container.Ports {
				g.Expect(port.HostPort).To(BeZero())
			}
			g.Expect(container.Command).To(ContainElement("--listen=" + values["kmsSocket"]))
			g.Expect(container.LivenessProbe.HTTPGet.Path).To(Equal("/healthz"))
			if controlPlaneOnly == "" {
				g.Expect(pod.NodeSelector).To(BeEmpty())
			} else {
				g.Expect(pod.NodeSelector).To(HaveKey("node-role.kubernetes.io/control-plane"))
			}
		})
	}
}

func validateKMSProviderIdentityToken(g Gomega, pod corev1.PodSpec, container corev1.Container, audience, tokenFile string) {
	var identityVolume *corev1.Volume
	var identityMount *corev1.VolumeMount
	for i := range pod.Volumes {
		volume := &pod.Volumes[i]
		if volume.Projected == nil {
			continue
		}
		for _, source := range volume.Projected.Sources {
			if source.ServiceAccountToken != nil {
				g.Expect(source.ServiceAccountToken.Audience).To(Equal(audience))
				g.Expect(source.ServiceAccountToken.Audience).To(Equal("sts.amazonaws.com"))
				g.Expect(source.ServiceAccountToken.ExpirationSeconds).To(HaveValue(Equal(int64(86400))))
				identityVolume = volume
				if mount := validateKMSProviderTokenMount(g, container, volume.Name, source.ServiceAccountToken.Path, tokenFile); mount != nil {
					identityMount = mount
				}
			}
		}
	}
	g.Expect(identityVolume).NotTo(BeNil())
	g.Expect(identityMount).NotTo(BeNil())
	g.Expect(identityVolume.Name).To(Equal("aws-iam-token"))
	g.Expect(identityMount.ReadOnly).To(BeTrue())
}

func validateKMSProviderTokenMount(g Gomega, container corev1.Container, volumeName, tokenPath, tokenFile string) *corev1.VolumeMount {
	var identityMount *corev1.VolumeMount
	for i := range container.VolumeMounts {
		mount := &container.VolumeMounts[i]
		if mount.Name == volumeName {
			identityMount = mount
			g.Expect(path.Join(mount.MountPath, tokenPath)).To(Equal(tokenFile))
		}
	}
	return identityMount
}
