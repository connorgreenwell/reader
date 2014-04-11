package main

import (
	"fmt"
	"github.com/SlyMarbo/rss"
	"time"
)

type Worker struct {
	URL      string
	Delay    time.Duration
	ItemChan chan *rss.Item
	QuitChan chan bool
}

func NewWorker(url string, items chan *rss.Item) Worker {
	worker := Worker{
		URL:      url,
		Delay:    1 * time.Second,
		ItemChan: items,
		QuitChan: make(chan bool),
	}
	return worker
}

func (w Worker) Start() {
	go func() {
		for {
			select {
			case <-time.After(w.Delay):
				fmt.Printf("fetcher(%s) fetching\n", w.URL)
				time.Sleep(10 * time.Second)
				feed, err := rss.Fetch(w.URL)
				w.Delay = feed.Refresh.Sub(time.Now())
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				for _, item := range feed.Items {
					w.ItemChan <- item
				}
				fmt.Printf("fetcher(%s) done\n", w.URL)
			case <-w.QuitChan:
				fmt.Printf("killing fetecher(%s)\n", w.URL)
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}
