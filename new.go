package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var newCmd = &Command{
	UsageLine: "new [sitename]",
	Short:     "create a site dir",
	Long: `
create a dir for site.

after then, you will edit the config.json file in .pd.
`,
}

var updateCmd = &Command{
	UsageLine: "update",
	Short:     "update site to new version",
	Long: `
update site .pd to new version.

after then, compile it and the site in new version.
`,
}

func init() {
	newCmd.Run = newApp
	updateCmd.Run = updateApp
	AddCommand(newCmd)
	AddCommand(updateCmd)
}

func newApp(cmd *Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "[ERRO] Argument [sitename] is missing")
		os.Exit(2)
	}
	sitedir := args[0]
	photo_dir := filepath.Join(sitedir, "photos", "thumb")
	err := os.MkdirAll(photo_dir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("create dir :%s.\n", photo_dir)
	video_dir := filepath.Join(sitedir, "videos")
	err = os.MkdirAll(video_dir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("create dir :%s.\n", video_dir)
	aboutfile := filepath.Join(sitedir, "about.md")
	ioutil.WriteFile(aboutfile, []byte("Hello, i am pd\n==="), os.ModePerm)
	fmt.Printf("create file :%s.\n", aboutfile)
	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(ZipInitStr))
	b, _ := ioutil.ReadAll(decoder)
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		log.Fatal(err)
	}
	for _, zf := range z.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		dst := filepath.Join(sitedir, zf.Name)
		os.MkdirAll(filepath.Dir(dst), os.ModePerm)
		f, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		rc, err := zf.Open()
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.Copy(f, rc)
		if err != nil {
			log.Fatal(err)
		}
		f.Sync()
		f.Close()
		rc.Close()
		fmt.Printf("create file :%s.\n", dst)
	}

	fmt.Printf("create site %s done.\n", sitedir)
}

func updateApp(cmd *Command, args []string) {
	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(ZipInitStr))
	b, _ := ioutil.ReadAll(decoder)
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		log.Fatal(err)
	}
	for _, zf := range z.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		if !strings.Contains(zf.Name, "/theme/") {
			continue
		}
		os.MkdirAll(filepath.Dir(zf.Name), os.ModePerm)
		f, err := os.OpenFile(zf.Name, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		rc, err := zf.Open()
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.Copy(f, rc)
		if err != nil {
			log.Fatal(err)
		}
		f.Sync()
		f.Close()
		rc.Close()
		fmt.Printf("update file :%s.\n", zf.Name)
	}

	fmt.Printf("update site done.\n")
}
