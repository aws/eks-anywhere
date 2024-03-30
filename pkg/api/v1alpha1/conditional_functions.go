package v1alpha1

// GetOrDefaultInt returns an non-zero int value or the default.
func GetOrDefaultInt(value, defaultValue int) int {
	if value != 0 {
		return value
	}

	return defaultValue
}

// GetOrDefaultOSFamily returns an non-empty OSFamily value or the default.
func GetOrDefaultOSFamily(value, defaultValue OSFamily) OSFamily {
	if value != "" {
		return value
	}

	return defaultValue
}

// GetOrDefaultString returns an non-empty string value or the default.
func GetOrDefaultString(value, defaultValue string) string {
	if value != "" {
		return value
	}

	return defaultValue
}

// GetOrDefaultStringArray returns an non-empty string value array or the default.
func GetOrDefaultStringArray(value, defaultValue []string) []string {
	if len(value) != 0 {
		return value
	}

	return defaultValue
}
