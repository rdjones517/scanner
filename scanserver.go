package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/pschou/go-exploder"
)

type ScanResult struct {
	FileName string `json:"filename"`
	Log      string `json:"log"`
	Result   string `json:"result"`
}

var (
	docRoot    = os.Args[1]
	explodeDir = path.Join(docRoot, "exploded")
	scanOpts   = []string{
		"--AFC=2048",
		"--MEMSIZE=5000000",
		"--THREADS=" + strconv.Itoa(runtime.NumCPU()),
		"--UNZIP",
		"--IGNORE-LINKS",
		"--MIME",
		"--RPTOBJECTS",
		"--SUMMARY",
	}
)

func explode(fileName string, file *os.File) {
	data := path.Join(explodeDir, fileName)
	log.Printf("Exploding file: %s into %s\n", fileName, data)
	stat, _ := file.Stat()
	log.Printf("file size: %d\n", stat.Size())
	file.Seek(0, io.SeekStart)
	os.MkdirAll(data, 0755)
	exploder.Explode(data, file, stat.Size(), -1)
}

func doScan(dir string) (result ScanResult, err error) {
	var res ScanResult
	var scanErr error

	log.Printf("scanning files in dir: %s", dir)

	args := append(scanOpts, dir)
	cmd := exec.Command("/usr/local/uvscan/uvscan", args...)

	out, err := cmd.Output()

	res.Log = string(out)

	if strings.Contains(res.Log, "No file or directory found") {
		scanErr = errors.New("file not found")
	}

	if err != nil {
		res.Result = "fail"
	} else {
		res.Result = "pass"
	}

	return res, scanErr

}

func cleanUp(fileName string, dir string) {
	log.Println("cleaning up.")
	os.Remove(fileName)
	os.RemoveAll(dir)
}

func Scan(w http.ResponseWriter, r *http.Request) {

	log.Println("receiving file")
	//write file to disk
	fileName := path.Join(docRoot, filepath.Base(r.URL.Path))
	dst, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = io.Copy(dst, r.Body)

	if err != nil {
		fmt.Println(err)
		return
	}

	//explode file
	explode(path.Base(fileName), dst)
	dst.Close()

	//doScan
	result, err := doScan(path.Join(explodeDir, fileName))
	result.FileName = path.Base(fileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

	cleanUp(fileName, path.Join(explodeDir, fileName))
}

func main() {
	// Simple static webserver:
	loadTLS()
	http.HandleFunc("/scan/", Scan)
	log.Println("Listening on HTTPS port 443 at /scan")
	server := &http.Server{Addr: ":443", TLSConfig: tlsConfig}
	log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
}
