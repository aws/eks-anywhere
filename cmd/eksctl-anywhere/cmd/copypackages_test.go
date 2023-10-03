package cmd

import (
	"net/http"
	"testing"

	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func TestGetTagsFromChartValues(t *testing.T) {
	res := make(map[string]string)
	chartValues := map[string]any{
		"controller1": map[string]any{
			"tag":    "testtag1",
			"digest": "testdiget1",
		},
		"controller2": map[string]any{
			"tag":    "testtag2",
			"digest": "testdiget2",
			"controller3": map[string]any{
				"tag":    "testtag3",
				"digest": "testdiget3",
			},
		},
	}
	err := getTagsFromChartValues(chartValues, res)
	if err != nil {
		t.Error(err)
	}
	if res["testdiget1"] != "testtag1" {
		t.Errorf("Expected tag has not be found")
	}
	if res["testdiget2"] != "testtag2" {
		t.Errorf("Expected tag has not be found")
	}
	if res["testdiget3"] != "testtag3" {
		t.Errorf("Expected tag has not be found")
	}
}

func TestGetPackageBundleTag(t *testing.T) {
	tag := getPackageBundleTag("1.27")
	if tag != "v1-27-latest" {
		t.Errorf("Expected tag v1-27-latest, got %s", tag)
	}
}

func TestSetupDstRepo(t *testing.T) {
	dst, err := remote.NewRepository("localhost:5000/hello-world")
	if err != nil {
		t.Error(err)
	}
	cpc := &copyPackagesConfig{dstPlainHTTP: true, dstInsecure: true}
	setUpDstRepo(dst, cpc)
	if dst.PlainHTTP != true {
		t.Errorf("Expect PlainHTTP to be true")
	}

	if dst.Client.(*auth.Client).Client.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify != true {
		t.Errorf("Expect InsecureSkipVerify to be true")
	}
}
