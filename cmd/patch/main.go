package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tiancaiamao/tidb-daily-bench"
)

func main() {
	origin, err := benchdaily.LoadDataDir("data")
	if err != nil {
		log.Fatal(err)
	}

	patch, err := benchdaily.LoadDataDir("patch")
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range patch {
		patched := false
		for _, o := range origin {
			if p.Date == o.Date && p.Commit == o.Commit {
				patchBenchOutput(p, o)
				patched = true
			}
		}

		if !patched {
			moveBenchOutput(p)
		}
	}
}

func moveBenchOutput(from benchdaily.BenchOutput) {
	tm, err := benchdaily.UnixDateToTime(from.Date)
	if err != nil {
		panic(err)
	}
	fileName := benchdaily.FileName(tm, from.Commit)
	os.Rename("patch/"+fileName, "data/"+fileName)
}

func patchBenchOutput(from, to benchdaily.BenchOutput) {
	for _, res := range from.Result {
		exist := false
		for _, v := range to.Result {
			if res.Name == v.Name {
				exist = true
			}
		}
		if exist {
			fmt.Printf("skip duplicated func %s in %s - %s\n", res.Name, from.Date, from.Commit)
			continue
		}

		to.Result = append(to.Result, res)
	}

	tm, err := benchdaily.UnixDateToTime(to.Date)
	if err != nil {
		panic(err)
	}

	outputFile := "data/" + benchdaily.FileName(tm, to.Commit)
	benchdaily.WriteJSONFile(outputFile, to)
}
