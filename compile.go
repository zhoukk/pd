package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var compileCmd = &Command{
	UsageLine: "compile config.json",
	Short:     "compile whole website",
	Long: `
	compile all markdown file in post.

	compile markdown file to html file.
	`,
}

var (
	Theme      string
	Root       string
	Config     Mapper
	Tpl        *template.Template
	Posts      AllPost
	Categories map[string]AllPost
	Videos     []string
	Photos     []string
)

type AllPost []Mapper

func (p AllPost) Len() int {
	return len(p)
}

func (p AllPost) Less(i, j int) bool {
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

func (p AllPost) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func init() {
	compileCmd.Run = compileApp
	AddCommand(compileCmd)
	log.SetFlags(log.Lshortfile)
	Categories = make(map[string]AllPost, 0)
}

func compileApp(cmd *Command, args []string) {
	config_file := "config.json"
	if len(args) > 0 {
		config_file = args[0]
	}
	err := LoadConf(config_file)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	err = LoadTheme()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	err = LoadPhotos()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	err = LoadVideos()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	err = LoadPosts()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	for i, v := range Posts {
		p, err := CreatePost(v, i)
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}
		Posts[i] = p
		if err := WritePostToFile(p); err != nil {
			log.Fatalln(err.Error())
			continue
		}
	}
	htmls := Config["htmls"].([]interface{})
	for _, v := range htmls {
		if err := CreateHtml(v.(string)); err != nil {
			log.Fatalln(err.Error())
		}
	}
	CopyJsCssImg()
	if err := CreateRss(); err != nil {
		log.Fatalln(err.Error())
	}
	if err := CreateAtom(); err != nil {
		log.Fatalln(err.Error())
	}
	if err := CreateSitemap(); err != nil {
		log.Fatalln(err.Error())
	}
}

func LoadConf(config_file string) error {
	f, err := os.Open(config_file)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&Config)
	if err != nil {
		return err
	}
	Theme = Config["theme"].(string)
	Root = Config["root"].(string)
	return nil
}

func LoadTheme() error {
	var tplfiles []string
	err := filepath.Walk(Theme+"/base/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") || !strings.HasSuffix(path, ".html") {
			return nil
		}
		tplfiles = append(tplfiles, path)
		return nil
	})
	if len(tplfiles) > 0 {
		Tpl = template.Must(template.ParseFiles(tplfiles...))
	}
	return err
}

func LoadPhotos() error {
	err := filepath.Walk(Root+"/image/thumb/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		Photos = append(Photos, filepath.Base(path))
		return nil
	})
	return err
}

func LoadVideos() error {
	err := filepath.Walk(Root+"/video/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		Videos = append(Videos, filepath.Base(path))
		return nil
	})
	return err
}

func LoadPosts() error {
	err := filepath.Walk(Root+"/post/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		post, err := LoadPost(path)
		if err != nil {
			return err
		}
		post["path"] = path
		category := post["category"].(string)
		if Categories[category] == nil {
			Categories[category] = make(AllPost, 0)
		}
		Categories[category] = append(Categories[category], post)
		Posts = append(Posts, post)
		return nil
	})

	sort.Sort(Posts)
	return err
}

func LoadPost(path string) (post Mapper, err error) {
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

func CreatePost(post Mapper, i int) (Mapper, error) {
	var prev, next *Mapper
	if i > 0 {
		next = &Posts[i-1]
	} else {
		next = nil
	}
	if i < len(Posts)-1 {
		prev = &Posts[i+1]
	} else {
		prev = nil
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
	return post, nil
}

func WritePostToFile(post Mapper) error {
	var buf bytes.Buffer
	link := filepath.Join(Root, post["permalink"].(string))
	err := os.MkdirAll(filepath.Dir(link), os.ModePerm)
	if err != nil {
		return err
	}
	var t *template.Template
	if Tpl != nil {
		t, _ = Tpl.Clone()
		t = template.Must(t.ParseFiles(Theme + "/post.html"))
	} else {
		t = template.Must(template.ParseFiles(Theme + "/post.html"))
	}
	err = t.Execute(&buf, Mapper{"categories": Categories, "post": post, "config": Config})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(link, buf.Bytes(), os.ModePerm)
	return err
}

func CreateHtml(html string) error {
	var buf bytes.Buffer
	var t *template.Template
	if Tpl != nil {
		t, _ = Tpl.Clone()
		t = template.Must(t.ParseFiles(Theme + html))
	} else {
		t = template.Must(template.ParseFiles(Theme + html))
	}
	err := t.Execute(&buf, Mapper{"categories": Categories, "posts": Posts, "photos": Photos, "videos": Videos, "config": Config})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(Root, html), buf.Bytes(), os.ModePerm)
	return err
}

func CopyJsCssImg() {
	CopyDir(filepath.Join(Theme, "js"), filepath.Join(Root, "/static/js"))
	CopyDir(filepath.Join(Theme, "css"), filepath.Join(Root, "/static/css"))
	CopyDir(filepath.Join(Theme, "img"), filepath.Join(Root, "/static/img"))
	CopyDir(filepath.Join(Theme, "fonts"), filepath.Join(Root, "/static/fonts"))
}

func CopyDir(srcpath, dstpath string) error {
	srcinfo, err := os.Stat(srcpath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dstpath, srcinfo.Mode())
	if err != nil {
		return err
	}
	dir, _ := os.Open(srcpath)
	objs, err := dir.Readdir(-1)
	for _, obj := range objs {
		srcfile := filepath.Join(srcpath, obj.Name())
		dstfile := filepath.Join(dstpath, obj.Name())

		if obj.IsDir() {
			err = CopyDir(srcfile, dstfile)
			if err != nil {
				log.Fatalln(err.Error())
			}
		} else {
			err = CopyFile(srcfile, dstfile)
			if err != nil {
				log.Fatalln(err.Error())
			}
		}
	}
	return err
}

func CopyFile(srcName, dstName string) error {
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err == nil {
		srcinfo, err := os.Stat(srcName)
		if err != nil {
			err = os.Chmod(dstName, srcinfo.Mode())
		}
	}
	return err
}
