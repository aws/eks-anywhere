package api

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestDeletePaths(t *testing.T) {
	tests := []struct {
		name         string
		obj, wantObj map[string]interface{}
		paths        []string
	}{
		{
			name: "delete nested paths, array and property from array of objects",
			obj: map[string]interface{}{
				"spec": map[string]interface{}{
					"workers": []interface{}{
						map[string]interface{}{
							"name":  "name1",
							"count": 1,
						},
						map[string]interface{}{
							"name":  "name2",
							"count": 2,
						},
					},
					"controlPlane": map[string]interface{}{
						"ip": "ip",
						"apiServer": map[string]interface{}{
							"cipher": "cipher1",
							"flags": []string{
								"flag1",
								"flag2",
							},
						},
					},
				},
			},
			paths: []string{
				"spec.controlPlane.apiServer.flags",
				"spec.workers[].name",
				"spec.controlPlane.ip",
			},
			wantObj: map[string]interface{}{
				"spec": map[string]interface{}{
					"workers": []interface{}{
						map[string]interface{}{
							"count": 1,
						},
						map[string]interface{}{
							"count": 2,
						},
					},
					"controlPlane": map[string]interface{}{
						"apiServer": map[string]interface{}{
							"cipher": "cipher1",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			deletePaths(tt.obj, tt.paths)
			g.Expect(tt.obj).To(Equal(tt.wantObj))
		})
	}
}
