package curatedpackages_test

import (
	"bytes"
	"encoding/json"
)

func convertJsonToBytes(obj interface{}) bytes.Buffer {
	b, _ := json.Marshal(obj)
	return *bytes.NewBuffer(b)
}
