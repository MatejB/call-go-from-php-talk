func chunks(fn string) chan []byte {
	chunkCh := make(chan []byte)

	go func() {
		defer close(chunkCh)

		handler := os.Open(fn)
		defer handler.Close()

		b := bytes.NewBuffer(make([]byte, 0, 1024))

		for {
			n := io.Copy(b, handler)
			if n == 0 {
				return
			}

			chunkCh <- b.Bytes()
			b.Reset()
		}
	}()

	return chunkCh
}
