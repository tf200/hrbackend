package ptr

func OrEmpty(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
