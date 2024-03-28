package api

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// ffmpeg -i "imagens-gratis.png" -q:v 1 -update true "imagens-gratis-out10.jpg"
// ffmpeg -i "imagens-gratis-out10.jpg" -filter:v scale="iw/1:ih/2" "imagens-gratis-out-out2.jpg"
// ffmpeg -i "imagens-gratis-out10.jpg" -filter:v scale=360:-2 "imagens-gratis-out-out2.jpg"
// ffmpeg -i "imagens-gratis-out10.jpg" -q:v 1 -update true "imagens-gratis-out11.jpg"

type ImageFile struct {
	Name     string
	Ext      string
	MimeType string
	Bytes    []byte
}

var (
	_, b, _, _ = runtime.Caller(0)
	RootPath   = filepath.Join(filepath.Dir(b), "../")
)

// convert png to jpg
func Convert(w http.ResponseWriter, r *http.Request) {

	// max total size 20mb
	r.ParseMultipartForm(200 << 20)

	// todo
	// handle file upload in bulk

	// fmt.Println(len(r.MultipartForm.File))
	// fmt.Println(len(r.MultipartForm.File["image"]))

	// for k, v := range r.MultipartForm.File["image"] {
	// 	fmt.Println(k)
	// 	fmt.Println(v.Filename)
	// 	fmt.Println(v.Size)
	// 	fmt.Println(v.Header)
	// }

	// validate file type

	// read file
	f, h, err := r.FormFile("image")
	if err != nil {
		fmt.Printf("Error reading file of 'image' form data. Reason: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filebytes, err := io.ReadAll(f)
	if err != nil {
		errStr := fmt.Sprintf("Error in reading the file buffer. Reason: %s\n", err)
		fmt.Println(errStr)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	srcImg := ImageFile{
		Name:     removedExt(h.Filename),
		Ext:      filepath.Ext(h.Filename),
		MimeType: h.Header.Get("Content-Type"),
		Bytes:    filebytes,
	}

	// options
	q := r.FormValue("q")
	if q == "" {
		q = "1"
	}

	// fmt.Printf("type: %s\n form : %s\n", reflect.TypeOf(opts["q"][0]), opts["q"][0])
	fmt.Printf("type: %s\n form : %s\n", reflect.TypeOf(q), q)

	buf, err := convertImg(&srcImg.Bytes, ffmpeg.KwArgs{
		"q:v": q,
		"c:v": "mjpeg",
	})

	if err != nil {
		fmt.Printf("failed processing image. Reason: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return converted
	w.Header().Set("Content-Type", srcImg.MimeType)
	w.Header().Set("Content-Disposition", `inline; filename="`+srcImg.Name+`.jpg"`)
	w.Write(buf.Bytes())
}

func Resize(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"message": "resize"}`))
}

func Compress(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"message": "compress"}`))
}

// utils

func removedExt(f string) string {
	return strings.TrimSuffix(f, filepath.Ext(f))
}

func convertImg(sourceImgBytes *[]byte, outputKwArgs ffmpeg.KwArgs) (*bytes.Buffer, error) {
	defOutputKwArgs := ffmpeg.KwArgs{
		"f": "image2",
		// "q:v":    "1",
		// "c:v":    "mjpeg",
		"update": "true",
	}

	maps.Copy(defOutputKwArgs, outputKwArgs)

	// process convert
	sourceImgBuf := bytes.NewBuffer(*sourceImgBytes)
	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input("pipe:").WithInput(sourceImgBuf).
		Output("pipe:", defOutputKwArgs).
		WithOutput(buf).
		Silent(true).
		Run()

	return buf, err
}

// func saveFile(f multipart.File, h *multipart.FileHeader) (ImageFile, error) {
// 	defer f.Close()

// 	tempFileName := fmt.Sprintf("uploaded-%s-*%s", removedExt(h.Filename), filepath.Ext(h.Filename))

// 	tempFile, err := os.CreateTemp(tempFolderPath, tempFileName)
// 	if err != nil {
// 		errStr := fmt.Sprintf("Error in creating the file %s\n", err)
// 		fmt.Println(errStr)
// 		return ImageFile{}, err
// 	}

// 	defer tempFile.Close()

// 	filebytes, err := io.ReadAll(f)
// 	if err != nil {
// 		errStr := fmt.Sprintf("Error in reading the file buffer %s\n", err)
// 		fmt.Println(errStr)
// 		return ImageFile{}, err
// 	}

// 	tempFile.Write(filebytes)

// 	_, tFilename := filepath.Split(tempFile.Name())

// 	imgFile := ImageFile{
// 		Name: tFilename,
// 		// FullPath: tempFile.Name(),
// 		MimeType: h.Header.Get("Content-Type"),
// 		Bytes:    filebytes,
// 	}

// 	return imgFile, nil
// }

// func removeFile(p string) bool {
// 	err := os.Remove(p)
// 	if err != nil {
// 		fmt.Printf("\nCannot remove file, : %s\n", err)
// 		return false
// 	}
// 	return true
// }
