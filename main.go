package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

var (
	data        []BenchOutput
	page        *components.Page
	pageVersion int64
	currVersion int64
)

type BenchOutput struct {
	Date   string
	Commit string
	Result []BenchResult
}

type BenchResult struct {
	Name        string
	NsPerOp     int64
	AllocsPerOp int64
	BytesPerOp  int64
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

func mainHandle(w http.ResponseWriter, _ *http.Request) {
	// if currVersion != pageVersion {
	final := groupByBench(data)
	page = makePage(final)
	pageVersion = currVersion
	// }
	page.Render(w)
}

func allocHandle(w http.ResponseWriter, _ *http.Request) {
	// if currVersion != pageVersion {
	final := groupByBench(data)
	page = makeAllocPage(final)
	pageVersion = currVersion
	// }
	page.Render(w)
}

func uploadHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method should be POST", http.StatusMethodNotAllowed)
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	var b BenchOutput
	err := dec.Decode(&b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
	}

	data = append(data, b)
	currVersion++
}

func loadDataDir() []BenchOutput {
	entries, err := os.ReadDir("data")
	if err != nil {
		panic(err)
	}

	res := make([]BenchOutput, 0, 100)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

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
		res = append(res, b)
	}
	return res
}

func groupByBench(entries []BenchOutput) map[string][]benchResult {
	final := make(map[string][]benchResult, len(entries))
	for _, b := range entries {
		v, err := strconv.ParseInt(b.Date, 10, 64)
		if err != nil {
			panic(err)
		}
		date := time.Unix(v, 0)
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

func makePage(final map[string][]benchResult) *components.Page {
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
			// allocs = append(allocs, v.AllocsPerOp)
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

func main() {
	data = loadDataDir()
	final := groupByBench(data)
	page = makePage(final)
	pageVersion = 0
	currVersion = 0

	http.HandleFunc("/", mainHandle)
	http.HandleFunc("/alloc", allocHandle)
	http.HandleFunc("/upload", uploadHandle)
	http.ListenAndServe(":18081", nil)
}
