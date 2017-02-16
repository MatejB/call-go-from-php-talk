package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/chrislusf/glow/driver"
	"github.com/chrislusf/glow/flow"
)

type DayHourResult struct {
	DayHour string
	Count   int
}

func init() {
	gob.Register(DayHourResult{})
}

func main() {
	flag.Parse()

	defer func(start time.Time) {
		fmt.Printf("\nElapsed: %v sec\n", time.Now().Sub(start).Seconds())
	}(time.Now())

	outCh := make(chan DayHourResult)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		data := [7][24]int{}

		included := 0
		var day, hour int
		for d := range outCh {
			fmt.Sscanf(d.DayHour, "%d-%d", &day, &hour)
			data[day][hour] += d.Count
			included += d.Count
		}

		maxGraphLength := float32(200)
		graphScale := float32(10)

		outDays := append(data[1:], data[:1]...)
		for d, hours := range outDays {
			switch d {
			case 0:
				fmt.Println("Monday")
			case 1:
				fmt.Println("Tuesday")
			case 2:
				fmt.Println("Wednesday")
			case 3:
				fmt.Println("Thursday")
			case 4:
				fmt.Println("Friday")
			case 5:
				fmt.Println("Saturday")
			case 6:
				fmt.Println("Sunday")
			}

			for h, visits := range hours {
				perc := float32(visits) / float32(included) * 100
				graph := maxGraphLength * (float32(visits) / float32(included)) * graphScale

				n := h + 1
				if n > 23 {
					n = 0
				}
				hour := fmt.Sprintf("%02d-%02d", h, n)

				gl := strconv.Itoa(int(graph))
				format := strings.Replace("\t%s\t%.4f\t%0[GL]s\n", "[GL]", gl, -1)

				fmt.Printf(format, hour, perc, "")
			}
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

func byDayAndHour(l string) (key string, value int) {
	p := strings.Split(l, " ")

	date := p[3]
	if len(date) > 0 && date[:1] == "[" {
		date = date[1:]
	}

	tz := p[4]
	if len(tz) > 0 && tz[len(tz)-1:] == "]" {
		tz = tz[:len(tz)-1]
	}

	t, err := time.Parse("2/Jan/2006:15:04:05 -0700", date+" "+tz)
	if err != nil {
		log.Fatal("Parse error ", err)
	}

	return fmt.Sprintf("%d-%d", t.Weekday(), t.Hour()), 1

}

func sum(a, b int) int {
	return a + b
}

func sendHome(dh string, count int, outputCh chan DayHourResult) {
	outputCh <- DayHourResult{dh, count}
}
