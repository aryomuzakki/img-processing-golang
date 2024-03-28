package api

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func TestConvertFunc(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
	}{
		{
			name:     "PNGTest1",
			filePath: "../assets/img/mock-data/image1.png",
		},
		{
			name:     "PNGTest2",
			filePath: "../assets/img/mock-data/image2.png",
		},
		{
			name:     "JPGTest3",
			filePath: "../assets/img/mock-data/image3.jpg",
		},
		{
			name:     "JPGTest4",
			filePath: "../assets/img/mock-data/image4.jpg",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			filebytes, err := os.ReadFile(test.filePath)
			require.Nilf(t, err, "Failed loading file: '%s'. Reason: %s\n", test.filePath, err)

			buf, err := convertImg(&filebytes, ffmpeg.KwArgs{
				"q:v": "1",
				"c:v": "mjpeg",
			})
			require.Nilf(t, err, "Failed converting file: '%s'. Reason: %s\n", test.filePath, err)
			require.Equal(t, "image/jpeg", http.DetectContentType(buf.Bytes()), "Result file type not jpg")

			// assert.Lessf(t, buf.Len(), len(filebytes), "File size not less for file: '%s'.", test.filePath)
			// fmt.Printf("%s: %s < %s\n", test.filePath, strconv.Itoa(buf.Len()), strconv.Itoa(len(filebytes)))

		})
	}

}

func TestConvertRoute(t *testing.T) {
	ct, body, err := formBody(map[string]string{
		"image": "file:../assets/img/mock-data/image3.jpg",
		"q":     "16",
	})
	require.Nilf(t, err, "Failed creating formdata for body. Reason: %s\n", err)

	r, err := http.NewRequest(http.MethodPost, "/convert", body)
	require.Nilf(t, err, "Failed creating http request. Reason: %s\n", err)
	r.Header.Set("Content-Type", ct)

	w := httptest.NewRecorder()

	Convert(w, r)

	resp := w.Result()

	respBody, err := io.ReadAll(resp.Body)
	require.Nilf(t, err, "Failed reading response body. Reason: %s\n", err)
	defer resp.Body.Close()

	fmt.Printf("length: %s\n", strconv.Itoa(len(respBody)))
	assert.Equal(t, resp.StatusCode, 200, "Expected 200 Response. Actual Status Code: %s\n", resp.StatusCode)

}

func formBody(data map[string]string) (string, io.Reader, error) {
	body := new(bytes.Buffer)

	mw := multipart.NewWriter(body)
	defer mw.Close()

	for key, val := range data {
		if strings.HasPrefix(val, "file:") {
			val = val[5:]
			fmt.Printf("val: %s", val)
			file, err := os.Open(val)
			if err != nil {
				return "", nil, err
			}
			defer file.Close()

			part, err := mw.CreateFormFile(key, val)
			if err != nil {
				return "", nil, err
			}
			io.Copy(part, file)
		} else {
			mw.WriteField(key, val)
		}
	}

	return mw.FormDataContentType(), body, nil
}
