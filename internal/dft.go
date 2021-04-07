package internal

func OrDefault(val, dft string) string {
	if val == "" {
		return dft
	}
	return val
}
