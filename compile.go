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
	InputPath  string
	OutputPath string
	Config     Mapper
	Tpl        *template.Template
	Posts      AllPost
	Tags       map[string]AllPost
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
	Tags = make(map[string]AllPost, 0)
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
	if err := CreateIndex(); err != nil {
		log.Fatalln(err.Error())
	}
	if err := CreateArchive(); err != nil {
		log.Fatalln(err.Error())
	}
	if err := CreateAbout(); err != nil {
		log.Fatalln(err.Error())
	}
	if err := CreateMsgBoard(); err != nil {
		log.Fatalln(err.Error())
	}
	if err := CreateAlbum(); err != nil {
		log.Fatalln(err.Error())
	}
	if err := CopyJsCssImg(); err != nil {
		log.Fatalln(err.Error())
	}
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
	InputPath = Config["input"].(string)
	OutputPath = Config["output"].(string)
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
	Tpl = template.Must(template.ParseFiles(tplfiles...))
	return err
}

func LoadPosts() error {
	err := filepath.Walk(InputPath, func(path string, info os.FileInfo, err error) error {
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
		tag := post["tags"].(string)
		if Tags[tag] == nil {
			Tags[tag] = make(AllPost, 0)
		}
		Tags[tag] = append(Tags[tag], post)
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
	link := filepath.Join(OutputPath, post["permalink"].(string))
	err := os.MkdirAll(filepath.Dir(link), os.ModePerm)
	if err != nil {
		return err
	}
	t, err := Tpl.Clone()
	t = template.Must(t.ParseFiles(Theme + "/post.html"))
	err = t.Execute(&buf, Mapper{"post": post, "config": Config})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(link, buf.Bytes(), os.ModePerm)
	return err
}

func CreateIndex() error {
	var buf bytes.Buffer
	t, err := Tpl.Clone()
	if err != nil {
		return err
	}
	t = template.Must(t.ParseFiles(Theme + "/index.html"))
	err = t.Execute(&buf, Mapper{"posts": Posts, "config": Config})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(OutputPath, "index.html"), buf.Bytes(), os.ModePerm)
	return err
}

func CreateArchive() error {
	var buf bytes.Buffer
	t, err := Tpl.Clone()
	if err != nil {
		return err
	}
	t = template.Must(t.ParseFiles(Theme + "/archive.html"))
	err = t.Execute(&buf, Mapper{"tags": Tags, "config": Config})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(OutputPath, "archive.html"), buf.Bytes(), os.ModePerm)
	return err
}

func CreateAbout() error {
	var buf bytes.Buffer
	t, err := Tpl.Clone()
	if err != nil {
		return err
	}
	t = template.Must(t.ParseFiles(Theme + "/about.html"))
	err = t.Execute(&buf, Mapper{"config": Config})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(OutputPath, "about.html"), buf.Bytes(), os.ModePerm)
	return err
}

func CreateMsgBoard() error {
	var buf bytes.Buffer
	t, err := Tpl.Clone()
	if err != nil {
		return err
	}
	t = template.Must(t.ParseFiles(Theme + "/msgboard.html"))
	err = t.Execute(&buf, Mapper{"config": Config})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(OutputPath, "msgboard.html"), buf.Bytes(), os.ModePerm)
	return err
}

func CreateAlbum() error {
	var buf bytes.Buffer
	t, err := Tpl.Clone()
	if err != nil {
		return err
	}
	t = template.Must(t.ParseFiles(Theme + "/album.html"))
	err = t.Execute(&buf, Mapper{"config": Config})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(OutputPath, "album.html"), buf.Bytes(), os.ModePerm)
	return err
}

func CopyJsCssImg() error {
	if err := CopyDir(filepath.Join(Theme, "js"), filepath.Join(OutputPath, "/static/js")); err != nil {
		return err
	}
	if err := CopyDir(filepath.Join(Theme, "css"), filepath.Join(OutputPath, "/static/css")); err != nil {
		return err
	}
	if err := CopyDir(filepath.Join(Theme, "img"), filepath.Join(OutputPath, "/static/img")); err != nil {
		return err
	}
	return nil
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
