package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var skipped, included int32
var days = [7][24]int32{}

func main() {
	http.HandleFunc("/", worker)

	// Good to know:
	// https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/
	s := &http.Server{
		Addr:           "127.0.0.1:9001",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}

func worker(w http.ResponseWriter, r *http.Request) {
	defer func(start time.Time) {
		fmt.Fprintf(w, "\nElapsed: %v sec\n", time.Now().Sub(start).Seconds())
	}(time.Now())

	chunkCh := chunks("resource/NASA_access_log_JulAug95")

	lineCh := lines(chunkCh)

	dataCh := extract(lineCh)

	sum(dataCh)

	maxGraphLength := float32(200)
	graphScale := float32(10)

	outDays := append(days[1:], days[:1]...)
	for d, hours := range outDays {
		switch d {
		case 0:
			fmt.Fprintln(w, "Monday")
		case 1:
			fmt.Fprintln(w, "Tuesday")
		case 2:
			fmt.Fprintln(w, "Wednesday")
		case 3:
			fmt.Fprintln(w, "Thursday")
		case 4:
			fmt.Fprintln(w, "Friday")
		case 5:
			fmt.Fprintln(w, "Saturday")
		case 6:
			fmt.Fprintln(w, "Sunday")
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

			fmt.Fprintf(w, format, hour, perc, "")
		}
	}

	fmt.Fprintf(w, "\nIncluded: %d\n", included)
	fmt.Fprintf(w, "Skipped: %d\n", skipped)
}

func chunks(fn string) chan []byte {
	chunkCh := make(chan []byte)

	go func() {
		defer close(chunkCh)

		handler, err := os.Open(fn)
		if err != nil {
			log.Fatal("File open error ", err)
		}
		defer handler.Close()

		b := bytes.NewBuffer(make([]byte, 0, 1024))
		for {
			n, err := io.Copy(b, handler)
			if err != nil {
				log.Fatal("Error in chucking ", err)
			}

			if n == 0 {
				return
			}

			chunkCh <- b.Bytes()
			b.Reset()
		}
	}()

	return chunkCh
}

func lines(chunkCh chan []byte) chan []byte {
	lineCh := make(chan []byte, 600)

	go func() {
		defer close(lineCh)

		var wg sync.WaitGroup

		wg.Add(4)
		for i := 0; i < 4; i++ {
			go func() {
				var oldC, sc, l []byte
				var nix int

				defer wg.Done()
			chunkLoop:
				for sc = range chunkCh {
					if len(oldC) > 0 {
						sc = append(oldC, sc...)
						oldC = nil
					}

					for len(sc) > 0 {
						i := 0
						for {
							i++
							nix = bytes.IndexByte(sc[i:], '\n')
							if nix == -1 {
								copy(oldC, sc)
								continue chunkLoop
							}
							break
						}

						l, sc = sc[:nix+1], sc[nix+2:]
						lineCh <- l
					}
				}
			}()
		}

		wg.Wait()
	}()

	return lineCh
}

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
					if len(l) == 0 {
						continue
					}

					p := bytes.Split(l, []byte(" "))

					if len(p) < 7 {
						log.Fatalf("Unexpected format %q.", l)
					}

					// filter only html pages
					url := p[6]
					if len(url) < 5 || string(url[len(url)-5:]) != ".html" {
						atomic.AddInt32(&skipped, 1)
						continue
					}

					atomic.AddInt32(&included, 1)

					// extract day in a week and hour in a day

					date := p[3]
					if len(date) > 0 && string(date[:1]) == "[" {
						date = date[1:]
					}

					tz := p[4]
					if len(tz) > 0 && string(tz[len(tz)-1:]) == "]" {
						tz = tz[:len(tz)-1]
					}

					t, err := time.Parse("2/Jan/2006:15:04:05 -0700", string(date)+" "+string(tz))
					if err != nil {
						log.Fatal("Parse error ", err)
					}

					dhCh <- dayHour{int(t.Weekday()), t.Hour()}
				}
			}()
		}

		wg.Wait()
	}()

	return dhCh
}

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
