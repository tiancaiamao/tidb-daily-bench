package main

import (
	"time"
	"fmt"
	"log"
	"strings"
	"strconv"
	"os/exec"
	"bytes"
)	

func main() {
	c := exec.Command("git", "log", "-n300", "--date=unix", "--pretty=format:%cd_%h")
	var out bytes.Buffer
	c.Stdout = &out
	err := c.Run()
	if err != nil {
		log.Println(err)
		return
	}

	var (
		lastYY int
		lastMM time.Month
		lastDD int
	)
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		tmp := strings.Split(line, "_")
		dateStr, githash := tmp[0], tmp[1]

		dateInt, err := strconv.ParseInt(dateStr, 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		date := time.Unix(dateInt, 0)
		yy, mm, dd := date.Date()

		if yy == lastYY && mm == lastMM && dd == lastDD {
			continue
		}
		
		checkout := exec.Command("git", "checkout", githash)
		checkout.Run()
		
		err = runCommand(dateStr, githash, date.Format("2006-01-02") + "_" + githash + ".json")
		if err != nil {
			log.Println("run command error", err)
			break
		}

		lastYY = yy
		lastMM = mm
		lastDD = dd
	}
}

func runCommand(unixDateStr, githash, outfile string) error {
	cmd := exec.Command("go", "test", "-run", "TestBenchDaily", "-date", unixDateStr, "-commit", githash, "-outfile", outfile)
	fmt.Println("running command ", cmd)
	return cmd.Run()
}
