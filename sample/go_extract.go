type dayHour struct {
	day  int
	hour int
}

func extract(lineCh chan []byte) chan dayHour {
	dhCh := make(chan dayHour, 10)

	go func() {
		defer close(dhCh)

		var wg sync.WaitGroup

		wg.Add(8)
		for i := 0; i < 8; i++ {
			go func() {
				defer wg.Done()

				for l := range lineCh {
					p := bytes.Split(l, []byte(" "))

					atomic.AddInt32(&included, 1)

					// ... date and tz extracted from p...

					t := time.Parse(...)

					dhCh <- dayHour{int(t.Weekday()), t.Hour()}
				}
			}()
		}

		wg.Wait()
	}()

	return dhCh
}
