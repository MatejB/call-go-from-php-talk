var days = [7][24]int32{}

func main() {
	chunkCh := chunks("resource/NASA_access_log_JulAug95")

	lineCh := lines(chunkCh)

	dataCh := extract(lineCh)

	sum(dataCh)

	// output

	fmt.Printf("Skipped: %d\n", skipped)
}
