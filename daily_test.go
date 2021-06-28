package session

import (
	"reflect"
	"runtime"
	"strings"
	// "fmt"
	"os"
	"testing"
	"encoding/json"
)

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

func TestXXX(t *testing.T) {
	tests := []func(b *testing.B){
		BenchmarkBasic,
		BenchmarkStringIndexLookup,
		BenchmarkTableScan,
	}
	
	res := make([]BenchResult, 0, len(tests))
	for _, t := range tests {
		name := callerName(t)
		// typ := reflect.TypeOf(t)
		r1 := testing.Benchmark(t)
		r2 := benchmarkResultToJson(name, r1)
		res = append(res, r2)
	}

	out, err := os.Create("xxx.out")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	err = enc.Encode(res)
	if err != nil {
		panic(err)
	}
}
