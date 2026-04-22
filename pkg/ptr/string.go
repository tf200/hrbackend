package ptr

func String(value string) *string {
	return &value
}

func OrEmpty(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
