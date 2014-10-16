package main

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type rss struct {
	Version string      `xml:"version,attr"`
	Channel *RssChannel `xml:"channel"`
}

type RssChannel struct {
	Title   string    `xml:"title"`
	Link    string    `xml:"link"`
	PubDate string    `xml:"pubDate"`
	Items   []RssItem `xml:"item"`
}

type RssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

func CreateRss() error {
	domain := Config["domain"].(string)
	items := make([]RssItem, 0)
	for _, v := range Posts {
		item := RssItem{v["title"].(string), domain + v["permalink"].(string), v["date"].(string), v["description"].(string)}
		items = append(items, item)
	}

	r := &rss{"2.0", &RssChannel{Config["title"].(string), domain, time.Now().Format(time.RFC822), items}}
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0"?>`)
	data, err := xml.Marshal(r)
	if err != nil {
		return err
	}
	buf.Write(data)
	err = ioutil.WriteFile(filepath.Join(OutputPath, "rss.xml"), buf.Bytes(), os.ModePerm)
	return err
}
