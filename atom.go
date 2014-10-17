package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type Feed struct {
	XMLName xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string   `xml:"title"`
	ID      string   `xml:"id"`
	Link    []Link   `xml:"link"`
	Updated string   `xml:"updated"`
	Author  *Person  `xml:"author"`
	Entry   []*Entry `xml:"entry"`
}

type Entry struct {
	Title     string  `xml:"title"`
	ID        string  `xml:"id"`
	Link      []Link  `xml:"link"`
	Published string  `xml:"published"`
	Updated   string  `xml:"updated"`
	Author    *Person `xml:"author"`
	Summary   *Text   `xml:"summary"`
	Content   *Text   `xml:"content"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type Person struct {
	Name     string `xml:"name"`
	URI      string `xml:"uri,omitempty"`
	Email    string `xml:"email,omitempty"`
	InnerXML string `xml:",innerxml"`
}

type Text struct {
	Type string `xml:"type,attr"`
	Body string `xml:",chardata"`
}

func CreateAtom() error {
	domain := Config["domain"].(string)
	feed := Feed{
		Title:   Config["title"].(string),
		ID:      domain,
		Updated: time.Now().Format(time.RFC822),
		Link: []Link{{
			Rel:  "self",
			Href: domain + "/atom.xml",
		}},
	}
	for _, v := range Posts {
		permalink := v["permalink"].(string)
		t, err := time.Parse("2006-01-02 15:04:05", v["date"].(string))
		if err != nil {
			t = time.Now()
		}
		cont := string(v["content"].(template.HTML))
		fmt.Println(cont)
		e := &Entry{
			Title: v["title"].(string),
			ID:    domain + permalink,
			Link: []Link{{
				Rel:  "alternate",
				Href: domain + permalink,
			}},
			Published: t.Format(time.RFC822),
			Updated:   t.Format(time.RFC822),
			Summary: &Text{
				Type: "html",
				Body: v["description"].(string),
			},
			Content: &Text{
				Type: "html",
				Body: cont,
			},
			Author: &Person{
				Name: Config["author"].(string),
			},
		}
		feed.Entry = append(feed.Entry, e)
	}
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0"?>`)
	data, err := xml.Marshal(feed)
	if err != nil {
		return err
	}
	buf.Write(data)
	err = ioutil.WriteFile(filepath.Join(OutputPath, "atom.xml"), buf.Bytes(), os.ModePerm)
	return err
}
