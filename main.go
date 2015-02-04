package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/fatih/color"
)

type item struct {
	Title string
	URL   string
}

func (i item) String() string {
	return fmt.Sprintf("%s\n[%s]", green(i.Title), yellow(i.URL))
}

type response struct {
	Data struct {
		Children []struct {
			Data item
		}
	}
}

var (
	subs     []string
	itemChan chan item
	wg       sync.WaitGroup
	done     bool
	start    time.Time
	subCount int
	numSubs  int
	red      = color.New(color.FgRed).SprintFunc()
	green    = color.New(color.FgYellow).SprintFunc()
	yellow   = color.New(color.FgGreen).SprintFunc()
)

func main() {
	nCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nCPU)

	subs = os.Args[1:]

	if len(subs) == 0 {
		subs = append(subs, "all")
	}

	itemChan := make(chan item)
	start = time.Now()
	numSubs = 0
	subCount = 0

	for i := 0; i < len(subs); i++ {
		wg.Add(1)
		go getSub(itemChan, subs[i])
	}

	for {
		if done {
			now := time.Now()
			diff := now.Sub(start) / time.Millisecond
			color.Cyan("Got: %s sub-reddits [total time: %s]\n", red(len(subs)), yellow(diff))
			break
		}

		select {
		case i := <-itemChan:
			subCount++
			if subCount >= numSubs {
				done = true
			}
			fmt.Printf("%s\n", i)
		}

	}
}

func getSub(itemChan chan item, sub string) <-chan item {
	defer wg.Done()

	u := fmt.Sprintf("http://reddit.com/r/%s.json", sub)
	color.Cyan("Fetching %s\n", u)

	resp, err := http.Get(u)
	if err != nil {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Error getting sub-reddit: %v\n", err)
			}
		}()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// done = true
		fmt.Printf("Error getting sub-reddit: %v\n", resp)
	}

	r := new(response)
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		fmt.Printf("Error parsing sub-reddit: %v\n", err)
	}

	numSubs += len(r.Data.Children)

	for _, child := range r.Data.Children {
		itemChan <- child.Data
	}

	return itemChan
}
