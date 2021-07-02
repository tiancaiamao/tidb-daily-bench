package main

import (
	"net/http"
	"sort"
	"strconv"
	"path"
	// "log"
	"os"
	"time"
	"strings"
	"encoding/json"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)


type BenchOutput struct {
	Date string
	Commit string
	Result []BenchResult
}

type BenchResult struct {
	Name string
	NsPerOp int64
	AllocsPerOp int64
	BytesPerOp int64
}

type benchResult struct {
	Date string
	Sort time.Time
	BenchResult 
}

type benchResultSlice []benchResult 

func (s benchResultSlice) Len() int {
	return len(s)
}

func (s benchResultSlice) Less(i, j int) bool {
	return s[i].Sort.Before(s[j].Sort)
}

func (s benchResultSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func httpserver(page *components.Page) func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		page.Render(w)
	}
}

func main() {
	entries, err := os.ReadDir("data")
	if err != nil {
		panic(err)
	}

	final := make(map[string][]benchResult, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		// strs := strings.Split(e.Name(), "_")
		// date, commit := strs[0], strs[1]
		// _ = commit

		f, err := os.Open(path.Join("data", e.Name()))
		if err != nil {
			panic(err)
		}
		defer f.Close()
		
		var b BenchOutput
		dec := json.NewDecoder(f)
		err = dec.Decode(&b)
		if err != nil {
			panic(err)
		}
		addToFinal(final, &b)
	}
	for _, v := range final {
		sort.Sort(benchResultSlice(v))
	}


	page := components.NewPage()
	for name, oneCase := range final {
		bar := charts.NewBar()
		bar.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{Title: name}))
		// charts.WithToolboxOpts(opts.{Show: true})

		dates := make([]string, 0, len(oneCase))
		nsop := make([]opts.BarData, 0, len(oneCase))
		// allocs := make([]int64, 0, len(oneCase))
		// byteAllocs := make([]int64, 0, len(oneCase))
		for _, v := range oneCase {
			dates = append(dates, v.Date)
			nsop = append(nsop, opts.BarData{Value: v.NsPerOp})
			// allocs = append(allocs, v.AllocsPerOp)
			// byteAllocs = append(byteAllocs, v.BytesPerOp)
		}

		bar.SetXAxis(dates)
		bar.AddSeries("ns/op", nsop)
		// bar.AddYAxis("allocs/op", allocs)
		// bar.AddYAxis("alloc bytes/op", byteAllocs)


		page.AddCharts(bar)
	}

	// f, err := os.Create("bar.html")
	// if err != nil {
	// 	log.Println(err)
	// }
	// page.Render(f)

	http.HandleFunc("/", httpserver(page))
	http.ListenAndServe(":18081", nil)
}

func addToFinal(final map[string][]benchResult, oneFile *BenchOutput) {
	v, err := strconv.ParseInt(oneFile.Date, 10, 64)
	if err != nil {
		panic(err)
	}
	date := time.Unix(v, 0)	
	for _, v := range oneFile.Result {
		benchCaseName := v.Name
		serialData, _ := final[benchCaseName]
		serialData = append(serialData, benchResult{
			Date: date.Format("2006-01-02"),
			Sort : date,
			BenchResult : v,		
		})
		final[benchCaseName] = serialData 
	}
}
