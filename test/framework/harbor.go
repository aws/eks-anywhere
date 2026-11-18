package framework

import (
	_ "embed"
	"fmt"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
)

//go:embed testdata/harbor_functional_check.sh
var harborFunctionalCheckScript string

const (
	harborAdminUsername = "admin"
	harborTestImageTag  = "v1"
)

// VerifyHarborFunctional checks the Harbor portal and API health, logs in as admin, and pushes and
// pulls back a small image from inside the cluster.
func (e *ClusterE2ETest) VerifyHarborFunctional(prefix, namespace string) {
	e.PushHarborTestImage(prefix, namespace, "library/eksa-e2e-"+utilrand.String(7))
}

// PushHarborTestImage pushes a small test image to repository (e.g. library/foo) and pulls it back.
// Use VerifyHarborTestImagePullable to check the image can still be pulled later.
func (e *ClusterE2ETest) PushHarborTestImage(prefix, namespace, repository string) {
	e.T.Logf("Pushing and pulling %s:%s in Harbor", repository, harborTestImageTag)
	e.runHarborCheck(prefix, namespace, repository, true)
}

// VerifyHarborTestImagePullable checks that an image pushed with PushHarborTestImage can still be pulled.
func (e *ClusterE2ETest) VerifyHarborTestImagePullable(prefix, namespace, repository string) {
	e.T.Logf("Pulling %s:%s from Harbor", repository, harborTestImageTag)
	e.runHarborCheck(prefix, namespace, repository, false)
}

func (e *ClusterE2ETest) runHarborCheck(prefix, namespace, repository string, push bool) {
	// The Harbor chart names its resources <release>-harbor-* unless the release name contains "harbor".
	coreSecret := fmt.Sprintf("%s-harbor-core", prefix)
	container := corev1.Container{
		Name:    "harbor-check",
		Image:   e.GetDiagnosticCollectorImage(),
		Command: []string{"bash", "-c", harborFunctionalCheckScript},
		Env: []corev1.EnvVar{
			// The chart's expose service (expose.nodePort.name) serves the portal, API and registry over TLS.
			{Name: "HARBOR_URL", Value: fmt.Sprintf("https://harbor.%s.svc.cluster.local", namespace)},
			{Name: "HARBOR_USERNAME", Value: harborAdminUsername},
			{
				// Use whatever admin password the install configured (the chart default unless the test
				// sets harborAdminPassword) without copying it into the test.
				Name: "HARBOR_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: coreSecret},
					Key:                  "HARBOR_ADMIN_PASSWORD",
				}},
			},
			{Name: "HARBOR_REPOSITORY", Value: repository},
			{Name: "HARBOR_TAG", Value: harborTestImageTag},
			{Name: "HARBOR_IMAGE_CONTENT", Value: "eks-a e2e " + repository},
			{Name: "HARBOR_PUSH", Value: strconv.FormatBool(push)},
		},
	}
	e.runCheckJob(namespace, "harbor-check", container, 15*time.Minute)
}
