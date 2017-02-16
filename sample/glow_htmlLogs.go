func htmlLogs(l string) bool {
	p := strings.Split(l, " ")

	if len(p) < 7 {
		log.Fatalf("Unexpected format %q.", l)
	}

	// filter only html pages
	url := p[6]
	if len(url) < 5 || url[len(url)-5:] != ".html" {
		return false
	}

	return true
}
