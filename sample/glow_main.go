import (
	"github.com/chrislusf/glow/flow"
)

type DayHourResult struct {
	DayHour string
	Count   int
}

func main() {
	outCh := make(chan DayHourResult)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		data := [7][24]int{}

		for d := range outCh {
			fmt.Sscanf(d.DayHour, "%d-%d", &day, &hour)
			data[day][hour] += d.Count
		}
	}()

	flow.New().
		TextFile("resource/NASA_access_log_JulAug95", 3).
		Filter(htmlLogs).
		Map(byDayAndHour).
		ReduceByKey(sum).
		Map(sendHome).
		AddOutput(outCh).
		Run()

	wg.Wait()
}
