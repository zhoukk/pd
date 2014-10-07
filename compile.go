package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var compileCmd = &Command{
	UsageLine: "compile",
	Short:     "compile whole website",
	Long: `
compile all markdown file in post.

compile markdown file to html file.
`,
}

var (
	theme  string
	config Mapper
	tpl    *template.Template
)

type Posts []Mapper

func (p Posts) Len() int {
	return len(p)
}

func (p Posts) Less(i, j int) bool {
	p1 := p[i]["date"].(string)
	p2 := p[j]["date"].(string)
	pt1, err := time.Parse("2006-01-02 15:04:05", p1)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	pt2, err := time.Parse("2006-01-02 15:04:05", p2)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	return pt1.After(pt2)
}

func (p Posts) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func init() {
	compileCmd.Run = compileApp
	AddCommand(compileCmd)
	log.SetFlags(log.Lshortfile)
}

func compileApp(cmd *Command, args []string) {
	err := loadConf()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	err = loadTheme()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	posts, err := loadPosts()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	var prev, next *Mapper
	for i, v := range posts {
		if i > 0 {
			next = &posts[i-1]
		} else {
			next = nil
		}
		if i < len(posts)-1 {
			prev = &posts[i+1]
		} else {
			prev = nil
		}
		p, err := createPost(v, prev, next)
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}
		posts[i] = p
	}
	if err := createIndex(posts); err != nil {
		log.Fatalln(err.Error())
	}
	if err := createArchive(posts); err != nil {
		log.Fatalln(err.Error())
	}
	if err := createAbout(); err != nil {
		log.Fatalln(err.Error())
	}
	if err := createMsgBoard(); err != nil {
		log.Fatalln(err.Error())
	}
}

func loadConf() error {
	f, err := os.Open("config.json")
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return err
	}
	theme = config["theme"].(string)
	return nil
}

func loadTheme() error {
	var tplfiles []string
	err := filepath.Walk("./theme/"+theme+"/base/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") || !strings.HasSuffix(path, ".html") {
			return nil
		}
		tplfiles = append(tplfiles, path)
		return nil
	})
	tpl = template.Must(template.ParseFiles(tplfiles...))
	return err
}

func loadPosts() (Posts, error) {
	posts := Posts{}

	err := filepath.Walk("./post/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		post, err := loadPost(path)
		if err != nil {
			return err
		}
		post["path"] = path
		posts = append(posts, post)
		return nil
	})

	sort.Sort(posts)
	return posts, err
}

func loadPost(path string) (post Mapper, err error) {
	post = make(Mapper, 0)
	content, err := ioutil.ReadFile(path)
	n := strings.IndexRune(string(content), '}')
	if n == -1 {
		err = errors.New("error format of post header")
		return
	}
	head := content[:n+1]
	err = json.Unmarshal(head, &post)
	if err != nil {
		return
	}
	post["content"] = string(content[n+1:])
	return
}

func createPost(post Mapper, prev, next *Mapper) (Mapper, error) {
	var buf bytes.Buffer
	link := "." + post["permalink"].(string)
	err := os.MkdirAll(filepath.Dir(link), os.ModePerm)
	if err != nil {
		return post, err
	}
	content := post["content"].(string)
	post["content"] = template.HTML(MarkdownToHtml(content))
	if prev != nil {
		post["previous_url"] = (*prev)["permalink"]
		post["previous_title"] = (*prev)["title"]
	}
	if next != nil {
		post["next_url"] = (*next)["permalink"]
		post["next_title"] = (*next)["title"]
	}
	ti, err := time.Parse("2006-01-02 15:04:05", post["date"].(string))
	if err != nil {
		return post, err
	}
	post["sdate"] = ti.Format("2006-01-02")
	t, err := tpl.Clone()
	if err != nil {
		return post, err
	}
	t = template.Must(t.ParseFiles("./theme/" + theme + "/post.html"))
	err = t.Execute(&buf, Mapper{"post": post, "config": config})
	if err != nil {
		return post, err
	}
	err = ioutil.WriteFile(link, buf.Bytes(), os.ModePerm)
	return post, err
}

func createIndex(posts Posts) error {
	var buf bytes.Buffer
	t, err := tpl.Clone()
	if err != nil {
		return err
	}
	t = template.Must(t.ParseFiles("./theme/" + theme + "/index.html"))
	err = t.Execute(&buf, Mapper{"posts": posts, "config": config})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./index.html", buf.Bytes(), os.ModePerm)
	return err
}

func createArchive(posts Posts) error {
	var buf bytes.Buffer
	t, err := tpl.Clone()
	if err != nil {
		return err
	}
	t = template.Must(t.ParseFiles("./theme/" + theme + "/archive.html"))
	err = t.Execute(&buf, Mapper{"posts": posts, "config": config})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./archive.html", buf.Bytes(), os.ModePerm)
	return err
}

func createAbout() error {
	var buf bytes.Buffer
	t, err := tpl.Clone()
	if err != nil {
		return err
	}
	t = template.Must(t.ParseFiles("./theme/" + theme + "/about.html"))
	err = t.Execute(&buf, Mapper{"config": config})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./about.html", buf.Bytes(), os.ModePerm)
	return err
}

func createMsgBoard() error {
	var buf bytes.Buffer
	t, err := tpl.Clone()
	if err != nil {
		return err
	}
	t = template.Must(t.ParseFiles("./theme/" + theme + "/msgboard.html"))
	err = t.Execute(&buf, Mapper{"config": config})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./msgboard.html", buf.Bytes(), os.ModePerm)
	return err
}
