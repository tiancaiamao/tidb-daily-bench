package main

import (
	"sort"
	"path"
	"log"
	"os"
	"time"
	"strings"
	"encoding/json"

	"github.com/go-echarts/go-echarts/charts"
)

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
		if !strings.HasSuffix(e.Name(), ".out") {
			continue
		}

		strs := strings.Split(e.Name(), "_")
		date, commit := strs[0], strs[1]
		_ = commit

		f, err := os.Open(path.Join("data", e.Name()))
		if err != nil {
			panic(err)
		}
		defer f.Close()
		
		var b []BenchResult
		dec := json.NewDecoder(f)
		err = dec.Decode(&b)
		if err != nil {
			panic(err)
		}
		addToFinal(final, date, b)
	}

	// var takeOne []benchResult
	// var title string
	// for name, oneCase := range final {
	// 	takeOne = oneCase
	// 	title = name
	// 	break
	// }

	title := "BenchmarkTableScan"
	takeOne  := final[title]

	sort.Sort(benchResultSlice(takeOne))


	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.TitleOpts{Title: title}, charts.ToolboxOpts{Show: true})

	dates := make([]string, 0, len(takeOne))
	nsop := make([]int64, 0, len(takeOne))
	// allocs := make([]int64, 0, len(takeOne))
	// byteAllocs := make([]int64, 0, len(takeOne))
	for _, v := range takeOne {
		dates = append(dates, v.Date)
		nsop = append(nsop, v.NsPerOp)
		// allocs = append(allocs, v.AllocsPerOp)
		// byteAllocs = append(byteAllocs, v.BytesPerOp)
	}

	bar.AddXAxis(dates)
	bar.AddYAxis("ns/op", nsop)
	// bar.AddYAxis("allocs/op", allocs)
	// bar.AddYAxis("alloc bytes/op", byteAllocs)

	f, err := os.Create("bar.html")
	if err != nil {
		log.Println(err)
	}
	bar.Render(f)
}

func addToFinal(final map[string][]benchResult, dateStr string, oneFile []BenchResult) {
	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		panic(err)
	}
	for _, v := range oneFile {
		benchCaseName := v.Name
		serialData, _ := final[benchCaseName]
		serialData = append(serialData, benchResult{
			Date: dateStr,
			Sort : date,
			BenchResult : v,		
		})
		final[benchCaseName] = serialData 
	}
}
