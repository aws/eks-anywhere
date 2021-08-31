package templater

const objectSeparator string = "\n---\n"

func AppendYamlResources(resources ...[]byte) []byte {
	separator := []byte(objectSeparator)

	size := 0
	for _, resource := range resources {
		size += len(resource) + len(separator)
	}

	b := make([]byte, 0, size)
	for _, resource := range resources {
		b = append(b, resource...)
		b = append(b, separator...)
	}

	return b
}
