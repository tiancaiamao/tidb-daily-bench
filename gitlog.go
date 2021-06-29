package main

import (
	"fmt"
	"log"
	"strings"
	"os/exec"
	"bytes"
)	

func main() {
	c := exec.Command("git", "log", "-n1000", "--date=short", "--pretty=format:%cd_%h")
	var out bytes.Buffer
	c.Stdout = &out
	err := c.Run()
	if err != nil {
		log.Println(err)
		return
	}

	var lastDate string
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		tmp := strings.Split(line, "_")
		date, githash := tmp[0], tmp[1]

		if date == lastDate {
			continue
		}
		
		checkout := exec.Command("git", "checkout", githash)
		checkout.Run()
		
		err := runCommand(date, githash)
		if err != nil {
			log.Println("run command error", err)
			break
		}
		lastDate = date
	}
}

func runCommand(date, githash string) error {
	cmd := exec.Command("go", "test", "-run", "TestXXX", "-date", date, "-commit", githash)
	fmt.Println("running command ", cmd)
	return cmd.Run()
}
