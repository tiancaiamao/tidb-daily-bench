package main

import (
	"encoding/json"
	"net/http"
	_ "net/http/pprof"
	"sort"
	"sync"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/tiancaiamao/tidb-daily-bench"
)

var (
	data      []benchdaily.BenchOutput
	allocPage *components.Page
	mainPage  *components.Page
	mu        sync.RWMutex
)

type benchResult struct {
	Date string
	Sort time.Time
	benchdaily.BenchResult
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

func mainHandle(w http.ResponseWriter, _ *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	mainPage.Render(w)
}

func allocHandle(w http.ResponseWriter, _ *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	allocPage.Render(w)
}

func uploadHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method should be POST", http.StatusMethodNotAllowed)
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	var b benchdaily.BenchOutput
	err := dec.Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	tm, err := benchdaily.UnixDateToTime(b.Date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	outfile := benchdaily.FileName(tm, b.Commit)
	err = benchdaily.WriteJSONFile("data/"+outfile, b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	data = append(data, b)
	mu.Unlock()
	reGeneratePage(data)
}

func groupByBench(entries []benchdaily.BenchOutput) map[string][]benchResult {
	final := make(map[string][]benchResult, len(entries))
	for _, b := range entries {
		date, err := benchdaily.UnixDateToTime(b.Date)
		if err != nil {
			panic(err)
		}
		for _, v := range b.Result {
			benchCaseName := v.Name
			serialData, _ := final[benchCaseName]
			serialData = append(serialData, benchResult{
				Date:        date.Format("2006-01-02"),
				Sort:        date,
				BenchResult: v,
			})
			final[benchCaseName] = serialData
		}
	}
	for _, v := range final {
		sort.Sort(benchResultSlice(v))
	}
	return final
}

func makeMainPage(final map[string][]benchResult) *components.Page {
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
	return page
}

func makeAllocPage(final map[string][]benchResult) *components.Page {
	page := components.NewPage()
	for name, oneCase := range final {
		bar := charts.NewBar()
		bar.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{Title: name}))
		// charts.WithToolboxOpts(opts.{Show: true})

		dates := make([]string, 0, len(oneCase))
		// nsop := make([]opts.BarData, 0, len(oneCase))
		allocs := make([]opts.BarData, 0, len(oneCase))
		// byteAllocs := make([]int64, 0, len(oneCase))
		for _, v := range oneCase {
			dates = append(dates, v.Date)
			allocs = append(allocs, opts.BarData{Value: v.AllocsPerOp})
			// byteAllocs = append(byteAllocs, v.BytesPerOp)
		}

		bar.SetXAxis(dates)
		bar.AddSeries("allocs/op", allocs)
		// bar.AddYAxis("allocs/op", allocs)
		// bar.AddYAxis("alloc bytes/op", byteAllocs)

		page.AddCharts(bar)
	}
	return page
}

func reGeneratePage(data []benchdaily.BenchOutput) {
	mu.RLock()
	final := groupByBench(data)
	tmpMainPage := makeMainPage(final)
	tmpAllocPage := makeAllocPage(final)
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()
	mainPage = tmpMainPage
	allocPage = tmpAllocPage
}

func main() {
	data, err := benchdaily.LoadDataDir("data")
	if err != nil {
		panic(err)
	}
	reGeneratePage(data)

	http.HandleFunc("/", mainHandle)
	http.HandleFunc("/alloc", allocHandle)
	http.HandleFunc("/upload", uploadHandle)
	http.ListenAndServe(":18081", nil)
}
