package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os/exec"
	"path"
	"strings"
)

var file string
var repo string
var host string

func init() {
	flag.StringVar(&file, "file", "", "file to upload symbols of")
	flag.StringVar(&repo, "repo", "", "the git repo the source file is part of")
	flag.StringVar(&host, "host", "", "host (with protocol) to upload symbols to")
}

func main() {
	flag.Parse()
	dsymutil := exec.Command("dsymutil", file)
	err := dsymutil.Run()
	if err != nil {
		log.Fatal(err)
	}

	dumpsyms := exec.Command("dump_syms", "-r", "-g", file+".dSYM", file)
	var in bytes.Buffer
	dumpsyms.Stdout = &in
	err = dumpsyms.Run()
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(in.Bytes()), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "FILE") {
			parts := strings.Split(line, " ")
			num := parts[1]
			opath := parts[2]
			if !strings.HasPrefix(opath, repo) {
				continue
			}
			opath = path.Join(path.Dir(opath), path.Base(opath))

			gitparts := strings.Split(repo, "/")
			origparts := strings.Split(opath, "/")
			npath := strings.Join(origparts[len(gitparts):], "/")

			lines[i] = fmt.Sprintf("FILE %s %s", num, npath)
		}
	}
	var out bytes.Buffer
	out.Write([]byte(strings.Join(lines, "\n")))
	body := upload(host, file, out)
	log.Printf("%s", string(body))
}

func upload(url, filename string, filedata bytes.Buffer) []byte {
	var b bytes.Buffer
	var err error
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("symfile", filename)
	if err != nil {
		log.Fatal(err)
	}
	if _, err = io.Copy(fw, &filedata); err != nil {
		log.Fatal(err)
	}
	w.Close()

	req, err := http.NewRequest("POST", url+"/symfiles", &b)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	var body bytes.Buffer
	io.Copy(&body, res.Body)
	return body.Bytes()
}
