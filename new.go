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
)

var newCmd = &Command{
	UsageLine: "new [sitename]",
	Short:     "create a site dir",
	Long: `
create a dir for site.

after then, you will edit the config.json file in .pd.
`,
}

func init() {
	newCmd.Run = newApp
	AddCommand(newCmd)
}

func newApp(cmd *Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "[ERRO] Argument [sitename] is missing")
		os.Exit(2)
	}
	sitedir := args[0]
	err := os.MkdirAll(filepath.Join(sitedir, "photos", "thumb"), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(filepath.Join(sitedir, "videos"), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

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
	}

	fmt.Printf("create site %s\n", sitedir)
}
