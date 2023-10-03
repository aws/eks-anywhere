package aflag

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Header Value.
type Header http.Header

// NewHeader returns a new Header pointer.
func NewHeader(h *http.Header) *Header {
	return (*Header)(h)
}

// String returns the string representation of the Header.
func (h *Header) String() string {
	if b, err := json.Marshal(h); err == nil {
		return string(b)
	}

	return ""
}

// Set sets the value of the Header.
// Format: "a=1;2,b=2;4;5".
func (h *Header) Set(val string) error {
	var ss []string
	n := strings.Count(val, "=")
	switch n {
	case 0:
		return fmt.Errorf("%s must be formatted as key=value;value", val)
	case 1:
		ss = append(ss, strings.Trim(val, `"`))
	default:
		r := csv.NewReader(strings.NewReader(val))
		var err error
		ss, err = r.Read()
		if err != nil {
			return err
		}
	}

	out := make(map[string][]string, len(ss))
	for _, pair := range ss {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("%s must be formatted as key=value;value", pair)
		}
		out[kv[0]] = strings.Split(kv[1], ";")
	}
	*h = out

	return nil
}

// Type returns the flag type.
func (h *Header) Type() string {
	return "header"
}
