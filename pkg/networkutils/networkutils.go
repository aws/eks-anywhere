package networkutils

import "strconv"

func IsPortValid(port string) bool {
	p, err := strconv.Atoi(port)
	return err == nil && p >= 1 && p <= 65535
}
