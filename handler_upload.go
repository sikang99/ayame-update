package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Set of media files for service
type MediaSet struct {
	SrcDir   string
	DstDir   string
	BaseDir  string
	Basename string     `json:"path_base,omitempty"`
	Files    []*os.File `json:"files,omitempty"`
	Desc     string     `json:"ops_cmd,omitempty"`
}

func (m *MediaSet) String() string {
	return fmt.Sprintf("MediaSet SrcDir: %s, DstDir: %s", m.SrcDir, m.DstDir)
}

// Handler for Uploading and Transcoding
func uploadHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	log.Printf("%s, %s", r.URL.Path, r.RemoteAddr)
	defer log.Printf("uploadHandler exit")

	// parse our multipart form, 10 << 20 specifies a maximum upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		log.Printf("FormFile error: %s", err)
		return
	}
	defer file.Close()

	// fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	// fmt.Printf("File Size: %+v\n", handler.Size)
	// fmt.Printf("MIME Header: %+v\n", handler.Header)

	// create a temp file within our upload directory that follows a particular naming pattern
	tempFile, err := ioutil.TempFile("upload", "cobot-*"+filepath.Ext(handler.Filename))
	if err != nil {
		log.Printf("TempFile error: %s", err)
		return
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("ReadAll error:", err)
		return
	}
	// write this byte array to our temporary file
	n, err := tempFile.Write(fileBytes)
	if int64(n) < handler.Size || err != nil {
		log.Println("Write error:", err, n)
		return
	}

	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully did upload the file and being processed it.\n")
	log.Printf("%s is stored to %s", handler.Filename, tempFile.Name())

	// prepare a media set for the upload file
	mset := &MediaSet{}
	mset.SrcDir = "upload/"
	mset.DstDir = "asset/record/"
	mset.BaseDir = time.Now().Format("D20060102T150405/")
	mset.Basename = filepath.Base(tempFile.Name())
	log.Println(mset)

	// do post media processing in background
	go postMediaProcessing(mset)
}

// Postprocessing for the video file uploaded
func postMediaProcessing(mset *MediaSet) (err error) {
	err = getMediaInfo(mset)
	if err != nil {
		log.Println("getMediaInfo error:", err)
		return
	}
	// log.Println("getMediaInfo:", "Done")

	err = makeMediaSet(mset)
	if err != nil {
		log.Println("makeMediaSet error:", err)
		return
	}
	// log.Println("makeMediaSet:", "Done")

	err = os.Remove(mset.SrcDir + mset.Basename)
	if err != nil {
		log.Println("Remove error:", err)
		return
	}
	return
}

// Make a set of media files for a video
func getMediaInfo(mset *MediaSet) (err error) {
	// check mediainfo command if it is executable
	path, err := exec.LookPath("mediainfo")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("mediainfo:", path)

	// Get meta information for the media file
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("mediainfo", mset.SrcDir+mset.Basename)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalln(err, string(stderr.Bytes()))
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	log.Println(outStr, errStr)
	return
}

// Make a set of media files for a video
func makeMediaSet(mset *MediaSet) (err error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("ffmpeg:", path)

	// generate related files for the input video
	os.MkdirAll(mset.DstDir+mset.BaseDir, os.ModePerm)
	// os.Chdir(mset.DstDir)

	inPart := mset.SrcDir + mset.Basename
	outPart := mset.DstDir + mset.BaseDir + mset.Basename
	log.Println(inPart, outPart)

	cmdStr := fmt.Sprintf("ffmpeg -y -i %s", inPart)

	mp4Opt := `-vf "scale=1280:720"`
	mp4Part := changePathExtention(outPart, ".mp4")
	cmdStr += fmt.Sprintf(" %s %s", mp4Opt, mp4Part)

	jpgOpt := `-ss 00:00:01.000 -frames:v 1 -vf "scale=640:360"`
	jpgPart := changePathExtention(outPart, ".jpg")
	cmdStr += fmt.Sprintf(" %s %s", jpgOpt, jpgPart)

	// gifOpt := `-r 10 -vf "scale=320:180" -loop 0`
	// gifPart := changePathExtention(outPart, ".gif")
	// cmdStr += fmt.Sprintf("%s %s", gifOpt, gifPart)

	// webpOpt := `-r 10 -vf "scale=320:180" -loop 0`
	// webpPart := changePathExtention(outPart, ".webp")
	// cmdStr += fmt.Sprintf(" %s %s", webpOpt, webpPart)

	// webmOpt := `-r 10 -vf "scale=320:180" -an`
	// webmPart := changePathExtention(outPart, ".webm")
	// cmdStr += fmt.Sprintf(" %s %s", webmOpt, webmPart)

	mpvOpt := `-r 10 -vf "scale=320:180" -an -f mp4`
	mpvPart := changePathExtention(outPart, ".mpv")
	cmdStr += fmt.Sprintf(" %s %s", mpvOpt, mpvPart)

	log.Println(cmdStr)
	cmd := exec.Command("sh", "-c", cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(out))
	return
}

// Change extension of the filename with another one
func changePathExtention(fpath, fext string) (str string) {
	ext := filepath.Ext(fpath)
	str = fpath[0:len(fpath)-len(ext)] + fext
	return
}

// Send a web page to the http client
func sendFilePage(w http.ResponseWriter, filename string) (err error) {
	page, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("ReadFile(%s) error: %s", filename, err)
		return
	}
	fmt.Fprintf(w, string(page))
	return
}

// newUUID generates a random UUID according to RFC 4122
func newUUIDString() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
