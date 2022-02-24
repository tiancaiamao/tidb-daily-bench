package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var ops string
var allocs string
var testName string

func usage() {
	fmt.Println(`Usage:
    bisect -bench BenchmarkXXX -ops low,high
    Or: bisect -bench BenchmarkYYY -allocs low,high`)
	os.Exit(-1)
}

func init() {
	flag.StringVar(&ops, "ops", "", "specify the ops range")
	flag.StringVar(&ops, "allocs", "", "specify the allocs range")
	flag.StringVar(&testName, "bench", "", "specify the bench function name")
}

func main() {
	flag.Parse()

	if testName == "" {
		usage()
	}
	opsFrom, opsTo := parseNumberPair(ops)
	if len(ops) > 0 && opsFrom >= opsTo {
		fmt.Println("ops from >= to", opsFrom, opsTo, ops)
		usage()
	}
	allocsFrom, allocsTo := parseNumberPair(allocs)
	if len(allocs) > 0 && allocsFrom >= allocsTo {
		fmt.Println("allocs from >= to", allocsFrom, allocsTo)
		usage()
	}

	var buf bytes.Buffer
	cmd := exec.Command("go", "test", "-benchmem", "-run", "XXX", "-bench", testName)
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	s := bufio.NewScanner(&buf)
	re0 := regexp.MustCompile("^Benchmark.*$")

	var str string
	for s.Scan() {
		// BenchmarkIntegerIndexScan-16    	    4896	    235082 ns/op	   82667 B/op	    1329 allocs/op
		str = re0.FindString(s.Text())
		if len(str) > 0 {
			break
		}
	}

	re1 := regexp.MustCompile("([0-9]+) ns/op")
	re2 := regexp.MustCompile("([0-9]+) B/op")
	re3 := regexp.MustCompile("([0-9]+) allocs/op")

	var NsPerOP, AllocsPerOP, BytesPerOP int64
	a := re1.FindStringSubmatch(str)
	if len(a) > 1 {
		NsPerOP, _ = strconv.ParseInt(a[1], 10, 64)
	}

	b := re2.FindStringSubmatch(str)
	if len(b) > 1 {
		AllocsPerOP, _ = strconv.ParseInt(b[1], 10, 64)
	}

	c := re3.FindStringSubmatch(str)
	if len(c) > 1 {
		BytesPerOP, _ = strconv.ParseInt(c[1], 10, 64)
	}

	if len(ops) > 0 && opsFrom > 0 && opsTo > 0 {
		// compare NsPerOP with [opsFrom, opsTo], and decide it's a good or bad case
		ret := goodOrBad(NsPerOP, opsFrom, opsTo)
		os.Exit(ret)
	}

	if len(allocs) > 0 && allocsFrom > 0 && allocsTo > 0 {
		// compare AllocsPerOP with [allocsFrom, allocsTo], and decide it's a good or bad case
		ret := goodOrBad(AllocsPerOP, allocsFrom, allocsTo)
		os.Exit(ret)
	}

	fmt.Println("AllocsPerOP:", AllocsPerOP)
	fmt.Println("NsPerOP:", NsPerOP)
	fmt.Println("BytesPerOP:", BytesPerOP)
}

func parseNumberPair(str string) (int64, int64) {
	tmp := strings.Split(str, ",")
	if len(tmp) == 2 {
		from, _ := strconv.ParseInt(tmp[0], 10, 64)
		to, _ := strconv.ParseInt(tmp[1], 10, 64)
		return from, to
	}
	return 0, 0
}

// Return 1~127 if the current source is bad (value near to to)
// Return 0 for a good case (val near to from)
func goodOrBad(val, from, to int64) int {
	if val > to {
		return 1
	}
	if val < from {
		return 0
	}

	if val > (from+to)/2 {
		return 1
	}
	return 0
}
