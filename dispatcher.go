package main

import (
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"fmt"
	"github.com/SlyMarbo/rss"
)

type Dispatcher struct {
	ItemChan  chan *rss.Item
	Workers   map[string]Worker
	StartChan chan string
	StopChan  chan string
	QuitChan  chan bool
	SQL       *sqlite3.Conn
}

func NewDispatcher(f string) (Dispatcher, error) {
	sql, err := sqlite3.Open(f)
	if err != nil {
		return Dispatcher{}, err
	}

	tables := map[string]bool{}
	row := make(sqlite3.RowMap)
	for s, err := sql.Query("SELECT name FROM sqlite_master WHERE type='table'"); err == nil; err = s.Next() {
		s.Scan(row)
		tables[row["name"].(string)] = true
	}

	if !tables["feeds"] {
		sql.Query("CREATE TABLE feeds (url TEXT, UNIQUE(url))")
	}

	if !tables["rss"] {
		sql.Query("CREATE TABLE rss (id TEXT, title TEXT, content TEXT, link TEXT, date DATETIME, UNIQUE(id))")
	}

	disp := Dispatcher{
		ItemChan:  make(chan *rss.Item),
		Workers:   make(map[string]Worker),
		StartChan: make(chan string),
		StopChan:  make(chan string),
		QuitChan:  make(chan bool),
		SQL:       sql,
	}
	return disp, nil
}

func (d Dispatcher) Start() {
	go func() {
		row := make(sqlite3.RowMap)
		for s, err := d.SQL.Query("SELECT url FROM feeds"); err == nil; err = s.Next() {
			s.Scan(row)
			d.StartChan <- row["url"].(string)
		}
	}()

	go func() {
		for item := range d.ItemChan {
			args := sqlite3.NamedArgs{
				"$id":      item.ID,
				"$title":   item.Title,
				"$content": item.Content,
				"$link":    item.Link,
				"$date":    item.Date,
			}
			d.SQL.Exec("INSERT INTO rss VALUES($id, $title, $content, $link, $date)", args)
		}
	}()

	go func() {
		for {
			select {
			case url := <-d.StartChan:
				fmt.Printf("starting worker for %s\n", url)
				d.Workers[url] = NewWorker(url, d.ItemChan)
				d.Workers[url].Start()
			case url := <-d.StopChan:
				fmt.Printf("halting worker for %s\n", url)
				d.Workers[url].Stop()
			case <-d.QuitChan:
				fmt.Println("halting dispatcher")
				return
			}
		}
	}()
}

func (d Dispatcher) Stop() {
	go func() {
		d.QuitChan <- true
	}()
}

func (d Dispatcher) StartFeed(url string) {
	go func() {
		args := sqlite3.NamedArgs{
			"$url": url,
		}
		d.SQL.Exec("INSERT INTO feeds VALUES($url)", args)
		d.StartChan <- url
		fmt.Println("Adding", url, "to feeds")
	}()
}

func (d Dispatcher) StopFeed(url string) {
	go func() {
		args := sqlite3.NamedArgs{
			"$url": url,
		}
		d.SQL.Exec("DELETE FROM feeds WHERE url=$url", args)
		d.StopChan <- url
		fmt.Println("Removing", url, "from feeds")
	}()
}
