package merger

func EnvMerge(dst, src map[string]string) {
	for key, value := range src {
		dst[key] = value
	}
}
