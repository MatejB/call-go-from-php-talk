func sum(dhCh chan dayHour) {
	var wg sync.WaitGroup

	wg.Add(6)
	for i := 0; i < 6; i++ {
		go func() {
			defer wg.Done()

			for d := range dhCh {
				atomic.AddInt32(&days[d.day][d.hour], 1)
			}
		}()
	}

	wg.Wait()
}
