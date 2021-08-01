package session

import (
	"testing"
	"runtime"
	"strings"
	"reflect"
	"flag"
	"fmt"
	"os"
	"context"
	"encoding/json"

	"github.com/pingcap/tidb/types"
)

func BenchmarkPointGet(b *testing.B) {
	ctx := context.Background()
	se, do, st := prepareBenchSession()
	defer func() {
		se.Close()
		do.Close()
		st.Close()
	}()
	mustExecute(se, "create table t (pk int primary key)")
	mustExecute(se, "insert t values (61),(62),(63),(64)")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute(ctx, "select * from t where pk = 64")
		if err != nil {
			b.Fatal(err)
		}
		_, err = drainRecordSet(ctx, se.(*session), rs[0])
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}

func BenchmarkBatchPointGet(b *testing.B) {
	ctx := context.Background()
	se, do, st := prepareBenchSession()
	defer func() {
		se.Close()
		do.Close()
		st.Close()
	}()
	mustExecute(se, "create table t (pk int primary key)")
	mustExecute(se, "insert t values (61),(62),(63),(64)")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute(ctx, "select * from t where pk in (61, 64, 67)")
		if err != nil {
			b.Fatal(err)
		}
		_, err = drainRecordSet(ctx, se.(*session), rs[0])
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}

func BenchmarkPreparedPointGet(b *testing.B) {
	ctx := context.Background()
	se, do, st := prepareBenchSession()
	defer func() {
		se.Close()
		do.Close()
		st.Close()
	}()
	mustExecute(se, "create table t (pk int primary key)")
	mustExecute(se, "insert t values (61),(62),(63),(64)")

	stmtID, _, _, err := se.PrepareStmt("select * from t where pk = ?")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.ExecutePreparedStmt(ctx, stmtID, []types.Datum{types.NewDatum(64)})
		if err != nil {
			b.Fatal(err)
		}
		_, err = drainRecordSet(ctx, se.(*session), rs)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}

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

func benchmarkResultToJSON(name string, r testing.BenchmarkResult) BenchResult {
	return BenchResult{
		Name:        name,
		NsPerOp:     r.NsPerOp(),
		AllocsPerOp: r.AllocsPerOp(),
		BytesPerOp:  r.AllocedBytesPerOp(),
	}
}

func callerName(f func(b *testing.B)) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	idx := strings.LastIndexByte(fullName, '.')
	if idx > 0 && idx+1 < len(fullName) {
		return fullName[idx+1:]
	}
	return fullName
}

var (
	date       = flag.String("date", "", " commit date")
	commitHash = flag.String("commit", "unknown", "brief git commit hash")
	outfile    = flag.String("outfile", "bench-daily.json", "specify the output file")
)

// TestBenchDaily collects the daily benchmark test result and generates a json output file.
// The format of the json output is described by the BenchOutput.
// Used by this command in the Makefile
// 	make bench-daily TO=xxx.json
func TestBenchDaily(t *testing.T) {
	if !flag.Parsed() {
		flag.Parse()
	}

	if *date == "" {
		// Don't run unless 'date' is specified.
		// Avoiding slow down the CI.
		return
	}

	tests := []func(b *testing.B){
		BenchmarkPreparedPointGet,
		BenchmarkPointGet,
		BenchmarkBatchPointGet,
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
		r1 := testing.Benchmark(t)
		r2 := benchmarkResultToJSON(name, r1)
		res = append(res, r2)
	}

	if *outfile == "" {
		*outfile = fmt.Sprintf("%s_%s.json", *date, *commitHash)
	}
	out, err := os.Create(*outfile)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()

	output := BenchOutput{
		Date:   *date,
		Commit: *commitHash,
		Result: res,
	}
	enc := json.NewEncoder(out)
	err = enc.Encode(output)
	if err != nil {
		t.Fatal(err)
	}
}
