package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"
)

// Trace represent one single log output.
type Trace struct {
	Id   string `json:"id"`
	Root Root   `json:"root"`
}

// Roots is an slice of root.
type Roots []Root

type Traces []Trace

// Root represent a struct of information that can handle one single log entry.
type Root struct {
	Service string `json:"service"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Calls   Roots  `json:"calls"`
	Span    string `json:"span"`
}

func main() {

	c := make(chan Trace)
	d := make(chan string)
	e := make(chan Trace)
	f := make(chan string)

	file, err := os.Create("logs.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go ReadLogs(c, d)
	go buildTrees(e, f, file)

	cOpen := true
	dOpen := true

	for cOpen || dOpen {
		select {
		case trace, open := <-c:
			if open {
				e <- trace
			} else {
				cOpen = false
			}
		case subTree, open := <-d:
			if open {
				f <- subTree
			} else {
				dOpen = false
			}
		}
	}
}

// ReadLogs read logs from the standard input
// The input is a stdin line by line.
// It accept 2 arguments as channel and provide them the function buildTraceRoot
// buildTraceRoot run as separate goroutine function.
func ReadLogs(c chan<- Trace, d chan<- string) {
	defer close(c)
	defer close(d)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading from input: ", err)
		}
		if line == "" {
			continue
		}
		go buildTraceRoot(line, c, d)
	}
}

// buildTraceRoot receive a single entry with 2 channels and a WaitGroup.
// c and d channels are non-blocking channels for write only
// For a single entry we test if there is a sub string "null" which means it's an entry service then new main root is created and sent in the channel c
// If the entry it's not an entry service then it's used for different process so the string is sent in th channel d for a different process.
func buildTraceRoot(entry string, c chan<- Trace, d chan<- string) {

	elements := strings.Fields(entry)
	if strings.Contains(entry, "null") {
		if len(elements) == 5 {
			var tr Trace
			root := newRoot(elements)
			tr.Id = elements[2]
			tr.Root = root
			c <- tr
		}
	} else {
		if len(elements) == 5 {
			d <- entry
		}
	}
}

// buildTrees append the traces inside a slice of traces and print it in json file.
// If the channel f receive a signal the we proceed for creating roots for every trace.
func buildTrees(e <-chan Trace, f <-chan string, file *os.File) {
	var traces []Trace
	for {
		select {
		case e := <-e:
			traces = append(traces, e)
		case f := <-f:
			elements := strings.Fields(f)

			buildRoots(elements, traces)
		}
		sortTraces(traces)
		for i := 0; i < len(traces); i++ {
			sortedCalls := sortCalls(traces[i].Root.Calls)
			traces[i].Root.Calls = sortedCalls

			jsonTrace, _ := json.Marshal(traces[i])
			fmt.Println(string(jsonTrace))
		}

		jsonTrace, _ := json.Marshal(traces)
		err := ioutil.WriteFile(file.Name(), jsonTrace, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}

	}

}

//buildRoots receive a slice of traces loop ever it and create a sub root for every trace based on the slice of stings provided.
// It checks if this entry elements belongs to it calls exists then it create a new root and push it inside the list of calls.
// The list of calls is sorted based on the start timestamps.
func buildRoots(elements []string, traces []Trace) {

	span := strings.Split(elements[4], "->")[0]

	for i := 0; i < len(traces); i++ {
		if traces[i].Root.Span == span {
			if len(traces[i].Root.Calls) > 0 {

				root := newRoot(elements)
				traces[i].Root.Calls = append(traces[i].Root.Calls, root)
				sortedCalls := sortCalls(traces[i].Root.Calls)
				traces[i].Root.Calls = sortedCalls

			} else if len(traces[i].Root.Calls) == 0 {
				root := newRoot(elements)
				traces[i].Root.Calls = append(traces[i].Root.Calls, root)

			}
		} else {
			calls := buildCalls(elements, traces[i].Root.Calls)
			sortedCalls := sortCalls(calls)
			traces[i].Root.Calls = sortedCalls
		}
	}

}

//buildCalls loop over a list roots and check if the element provided could belongs to this list or one of the sub roots of this list then create this call and push it inside.
//buildCalls calls itself recursively and build nested calls.
func buildCalls(elements []string, calls []Root) []Root {

	span := strings.Split(elements[4], "->")[0]

	for i := 0; i < len(calls); i++ {
		if calls[i].Span == span {
			if len(calls[i].Calls) == 0 {
				root := newRoot(elements)
				calls[i].Calls = append(calls[i].Calls, root)
			} else {
				root := newRoot(elements)
				calls[i].Calls = append(calls[i].Calls, root)
				sortedCalls := sortCalls(calls[i].Calls)
				calls[i].Calls = sortedCalls
			}
		} else {
			buildCalls(elements, calls[i].Calls)
		}
	}

	return calls
}

func newRoot(elements []string) Root {
	return Root{
		Service: elements[3],
		Start:   elements[0],
		End:     elements[1],
		Span:    strings.Split(elements[4], "->")[1],
		Calls:   Roots{},
	}
}

func (p Roots) Len() int {
	return len(p)
}

func (p Roots) Less(i, j int) bool {
	start, _ := time.Parse(time.RFC3339, p[i].Start)
	nextStart, _ := time.Parse(time.RFC3339, p[j].Start)
	return start.Before(nextStart)
}

func (p Roots) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func sortCalls(roots Roots) Roots {
	sort.Sort(Roots(roots))
	return roots
}

func sortTraces(traces Traces) Traces {
	sort.Sort(Traces(traces))
	return traces
}

func (p Traces) Len() int {
	return len(p)
}

func (p Traces) Less(i, j int) bool {
	start, _ := time.Parse(time.RFC3339, p[i].Root.Start)
	nextStart, _ := time.Parse(time.RFC3339, p[j].Root.Start)
	return start.Before(nextStart)
}

func (p Traces) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
