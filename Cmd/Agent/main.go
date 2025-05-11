package main

import (
	. "Calculator3.0/Internal/Agent"
	"os"
	"strconv"
	"sync"
)

func main() {
	numAgents, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if numAgents <= 0 {
		numAgents = 2 // По умолчанию 2 агента
	}

	wg := &sync.WaitGroup{}

	for i := 1; i <= numAgents; i++ {
		agent := NewAgent(i, wg)
		wg.Add(1)
		go agent.Run()
	}

	wg.Wait()
}
