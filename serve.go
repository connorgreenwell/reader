package main

import (
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Item struct {
	ID      string
	Title   string
	Link    string
	Content template.HTML
	Date    time.Time
}

type Page struct {
	Items   []Item
	Prev    string
	Next    string
	HasPrev bool
	HasNext bool
	End     bool
	Feeds   []string
}

func getItems(page, size int) []Item {
	var items []Item

	sql, _ := sqlite3.Open("rss.db")
	row := make(sqlite3.RowMap)
	for s, err := sql.Query(fmt.Sprintf("SELECT * FROM rss ORDER BY date DESC LIMIT %d OFFSET %d", size, page*size)); err == nil; err = s.Next() {
		s.Scan(row)
		item := Item{
			ID:      row["id"].(string),
			Title:   row["title"].(string),
			Content: template.HTML(row["content"].(string)),
			Link:    row["link"].(string),
			Date:    row["date"].(time.Time),
		}
		items = append(items, item)
	}

	return items
}

func getFeeds() []string {
	var feeds []string

	sql, _ := sqlite3.Open("rss.db")
	row := make(sqlite3.RowMap)
	for s, err := sql.Query("SELECT url FROM feeds"); err == nil; err = s.Next() {
		s.Scan(row)
		feed, _ := url.QueryUnescape(row["url"].(string))
		feeds = append(feeds, feed)
	}

	return feeds
}

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	page := 0
	size := 10

	v := r.URL.Query()
	if len(v["page"]) > 0 {
		p, err := strconv.Atoi(v["page"][0])
		if err == nil {
			page = p
		}
	}
	if len(v["size"]) > 0 {
		s, err := strconv.Atoi(v["size"][0])
		if err == nil {
			size = s
		}
	}

	items := getItems(page, size)

	p := Page{
		Items:   items,
		Next:    fmt.Sprintf("/?page=%d&size=%d", page+1, size),
		Prev:    fmt.Sprintf("/?page=%d&size=%d", page-1, size),
		HasPrev: page != 0,
		HasNext: !(len(items) < size),
		End:     (len(items) < size),
		Feeds:   getFeeds(),
	}

	t, err := template.ParseFiles("page.html")
	if err != nil {
		fmt.Println(err)
	}
	t.Execute(w, p)
}

func AddHandler(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()["blog"][0]
	disp.StartFeed(v)
	http.Redirect(w, r, "/", http.StatusFound)
}

func RemoveHandler(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()["blog"][0]
	disp.StopFeed(v)
	http.Redirect(w, r, "/", http.StatusFound)
}
