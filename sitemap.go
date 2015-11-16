package main

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"os"
	"time"
)

type urlset struct {
	Xmlns string       `xml:"xmlns,attr"`
	Urls  []SitemapUrl `xml:"url"`
}

type SitemapUrl struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq"`
	Priority   string `xml:"priority"`
}

func CreateSitemap() error {
	domain := Config["domain"].(string)
	urls := make([]SitemapUrl, 0)
	for _, v := range Posts {
		item := SitemapUrl{domain + v["permalink"].(string), time.Now().Format("2006-01-02"), "weekly", "1.0"}
		urls = append(urls, item)
	}

	r := &urlset{"http://www.sitemaps.org/schemas/sitemap/0.9", urls}
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	data, err := xml.Marshal(r)
	if err != nil {
		return err
	}
	buf.Write(data)
	err = ioutil.WriteFile("sitemap.xml", buf.Bytes(), os.ModePerm)
	return err
}
