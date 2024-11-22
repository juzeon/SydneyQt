package util

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"os"
	"strings"
	"time"
)

type ImageUploadResult struct {
	Base64URL string `json:"base64_url"`
	URL       string `json:"bing_url"`
	Canceled  bool   `json:"canceled"`
}
type ImageUploader interface {
	UploadFromBase64(rawBase64 string) (ImageUploadResult, error)
	UploadByFileSelector() (ImageUploadResult, error)
}
type CatboxImageUploader struct {
	Proxy string
	Ctx   context.Context
}

func (o CatboxImageUploader) UploadFromBase64(rawBase64 string) (ImageUploadResult, error) {
	var empty ImageUploadResult
	v, err := base64.StdEncoding.DecodeString(rawBase64)
	if err != nil {
		return empty, err
	}
	return o.uploadBytes(v)
}

func (o CatboxImageUploader) UploadByFileSelector() (ImageUploadResult, error) {
	var empty ImageUploadResult
	file, err := runtime.OpenFileDialog(o.Ctx, runtime.OpenDialogOptions{
		Title: "Open an image to upload",
		Filters: []runtime.FileFilter{{
			DisplayName: "Image Files (*.jpg; *.jpeg; *.png; *.gif)",
			Pattern:     "*.jpg;*.jpeg;*.png;*.gif",
		}},
	})
	if err != nil {
		return empty, err
	}
	if file == "" {
		return ImageUploadResult{Canceled: true}, nil
	}
	v, err := os.ReadFile(file)
	if err != nil {
		return empty, err
	}
	return o.uploadBytes(v)
}
func (o CatboxImageUploader) uploadBytes(v []byte) (ImageUploadResult, error) {
	var empty ImageUploadResult
	jpgData, err := ConvertImageToJpg(v)
	if err != nil {
		return empty, err
	}
	_, client, err := MakeHTTPClient(o.Proxy, 30*time.Second)
	if err != nil {
		return empty, err
	}
	resp, err := client.R().SetFormData(map[string]string{
		"reqtype":  "fileupload",
		"userhash": "",
	}).SetFileBytes("fileToUpload", "a.jpg", jpgData).Post("https://catbox.moe/user/api.php")
	if err != nil {
		return empty, err
	}
	if resp.IsErrorState() {
		return empty, errors.New("upload image failed: " + resp.GetStatus())
	}
	return ImageUploadResult{
		Base64URL: "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(jpgData),
		URL:       strings.TrimSpace(resp.String()),
	}, nil
}

type SydneyImageUploader struct {
	SydneyUploadImage func(jpgImgData []byte) (string, error)
	Ctx               context.Context
}

func (o SydneyImageUploader) UploadFromBase64(rawBase64 string) (ImageUploadResult, error) {
	var empty ImageUploadResult
	v, err := base64.StdEncoding.DecodeString(rawBase64)
	if err != nil {
		return empty, err
	}
	jpgData, err := ConvertImageToJpg(v)
	if err != nil {
		return empty, err
	}
	url, err := o.SydneyUploadImage(jpgData)
	if err != nil {
		return empty, err
	}
	return ImageUploadResult{
		Base64URL: "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(jpgData),
		URL:       url,
	}, err
}

func (o SydneyImageUploader) UploadByFileSelector() (ImageUploadResult, error) {
	var empty ImageUploadResult
	file, err := runtime.OpenFileDialog(o.Ctx, runtime.OpenDialogOptions{
		Title: "Open an image to upload",
		Filters: []runtime.FileFilter{{
			DisplayName: "Image Files (*.jpg; *.jpeg; *.png; *.gif)",
			Pattern:     "*.jpg;*.jpeg;*.png;*.gif",
		}},
	})
	if err != nil {
		return empty, err
	}
	if file == "" {
		return ImageUploadResult{Canceled: true}, nil
	}
	v, err := os.ReadFile(file)
	if err != nil {
		return empty, err
	}
	jpgData, err := ConvertImageToJpg(v)
	if err != nil {
		return empty, err
	}
	url, err := o.SydneyUploadImage(jpgData)
	if err != nil {
		return empty, err
	}
	return ImageUploadResult{
		Base64URL: "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(jpgData),
		URL:       url,
	}, err
}
