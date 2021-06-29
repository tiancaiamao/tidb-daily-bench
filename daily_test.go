package session

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"flag"
	"os"
	"testing"
	"encoding/json"
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

func benchmarkResultToJson(name string, r testing.BenchmarkResult) BenchResult {
	return BenchResult{
		Name: name,
		NsPerOp: r.NsPerOp(),
		AllocsPerOp: r.AllocsPerOp(),
		BytesPerOp : r.AllocedBytesPerOp(),
	}
}

func callerName(f func(b *testing.B)) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	idx := strings.LastIndexByte(fullName, '.')
	if idx > 0 && idx+1 < len(fullName){
		return fullName[idx+1:]
	}
	return fullName
}

var (
	date = flag.String("date", "2021-05-06", "date of this commit")
	commitHash = flag.String("commit", "0ec8f2d9f", "brief git commit hash")
)

func TestXXX(t *testing.T) {
	if !flag.Parsed() {
		flag.Parse()
	}

	tests := []func(b *testing.B){
		BenchmarkBasic,
		BenchmarkTableScan,
		BenchmarkTableLookup,
		BenchmarkExplainTableLookup,
		BenchmarkStringIndexScan,
		BenchmarkExplainStringIndexScan,
		BenchmarkStringIndexLookup,
		BenchmarkIntegerIndexScan,
		BenchmarkIntegerIndexLookup,
		BenchmarkDecimalIndexScan,
		BenchmarkDecimalIndexLookup,
		BenchmarkInsertWithIndex,
		BenchmarkInsertNoIndex,
		BenchmarkSort,
		BenchmarkJoin,
		BenchmarkJoinLimit,
		BenchmarkPartitionPruning,
		BenchmarkRangeColumnPartitionPruning,
		BenchmarkHashPartitionPruningPointSelect,
		BenchmarkHashPartitionPruningMultiSelect,
	}
	
	res := make([]BenchResult, 0, len(tests))
	for _, t := range tests {
		name := callerName(t)
		// typ := reflect.TypeOf(t)
		r1 := testing.Benchmark(t)
		r2 := benchmarkResultToJson(name, r1)
		res = append(res, r2)
	}

	out, err := os.Create(fmt.Sprintf("%s_%s.json", *date, *commitHash))
	if err != nil {
		panic(err)
	}
	defer out.Close()

	output := BenchOutput{
		Date : *date,
		Commit : *commitHash,
		Result: res,
	}
	enc := json.NewEncoder(out)
	err = enc.Encode(output)
	if err != nil {
		panic(err)
	}
}
