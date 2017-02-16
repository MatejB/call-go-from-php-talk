func main() {
	http.HandleFunc("/", worker)

	s := &http.Server{
		Addr: "127.0.0.1:9001",
	}

	s.ListenAndServe()
}

func worker(w http.ResponseWriter, r *http.Request) {
	chunkCh := chunks("resource/NASA_access_log_JulAug95")

	lineCh := lines(chunkCh)

	dataCh := extract(lineCh)

	sum(dataCh)

	// ....

	fmt.Fprintf(w, "Skipped: %d\n", skipped)
}
