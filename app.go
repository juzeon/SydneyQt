package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/flytam/filenamify"
	goversion "github.com/hashicorp/go-version"
	"github.com/pkoukk/tiktoken-go"
	"github.com/samber/lo"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sydneyqt/sydney"
	"sydneyqt/util"
	"sync"
	"time"
)

//go:embed version.txt
var version string

// App struct
type App struct {
	settings *Settings
	ctx      context.Context
	logFile  *os.File
	logToStd bool
}

// NewApp creates a new App application struct
func NewApp(settings *Settings) *App {
	return &App{settings: settings}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	env := runtime.Environment(ctx)
	a.logFile = os.Stderr
	a.logToStd = true
	if env.BuildType == "production" {
		f, err := os.OpenFile(util.WithPath("log_"+time.Now().Format("2006-01")+".log"),
			os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			util.GracefulPanic(err)
		}
		a.logFile = f
		a.logToStd = false
	}
	a.updateLogger(a.settings.config.Debug)
	go func() {
		for debug := range a.settings.DebugChangeSignal {
			a.updateLogger(debug)
		}
	}()
}
func (a *App) shutdown(ctx context.Context) {
	if !a.logToStd {
		a.logFile.Close()
	}
	a.settings.Exit <- struct{}{}
	os.Exit(0)
}
func (a *App) updateLogger(debug bool) {
	slog.SetDefault(slog.New(slog.NewTextHandler(a.logFile, &slog.HandlerOptions{
		AddSource: true,
		Level:     lo.Ternary(debug, slog.LevelDebug, slog.LevelInfo),
	})))
	slog.Info("Update logger", "debug", debug)
}

var tk *tiktoken.Tiktoken
var initTkFunc = sync.OnceFunc(func() {
	slog.Info("Init tiktoken")
	t, err := tiktoken.EncodingForModel("gpt-4")
	if err != nil {
		util.GracefulPanic(err)
	}
	tk = t
})

func (a *App) CountToken(text string) int {
	initTkFunc()
	return len(tk.Encode(text, nil, nil))
}

type UploadSydneyImageResult struct {
	Base64URL string `json:"base64_url"`
	BingURL   string `json:"bing_url"`
	Canceled  bool   `json:"canceled"`
}

func (a *App) getImageUploader() (util.ImageUploader, error) {
	var empty util.ImageUploader
	workspace, err := a.settings.config.GetCurrentWorkspace()
	if err != nil {
		return empty, err
	}
	if workspace.Backend == "Sydney" {
		sydneyIns, err := a.createSydney()
		if err != nil {
			return empty, err
		}
		return util.SydneyImageUploader{
			SydneyUploadImage: sydneyIns.UploadImage,
			Ctx:               a.ctx,
		}, nil
	} else {
		return util.CatboxImageUploader{
			Proxy: a.settings.config.Proxy,
			Ctx:   a.ctx,
		}, nil
	}
}
func (a *App) UploadSydneyImageFromBase64(rawBase64 string) (UploadSydneyImageResult, error) {
	var empty UploadSydneyImageResult
	imageUploader, err := a.getImageUploader()
	if err != nil {
		return empty, err
	}
	res, err := imageUploader.UploadFromBase64(rawBase64)
	if err != nil {
		return empty, err
	}
	return UploadSydneyImageResult{
		Base64URL: res.Base64URL,
		BingURL:   res.URL,
	}, nil
}
func (a *App) UploadSydneyImage() (UploadSydneyImageResult, error) {
	var empty UploadSydneyImageResult
	imageUploader, err := a.getImageUploader()
	if err != nil {
		return empty, err
	}
	res, err := imageUploader.UploadByFileSelector()
	if err != nil {
		return empty, err
	}
	return UploadSydneyImageResult{
		Base64URL: res.Base64URL,
		BingURL:   res.URL,
	}, nil
}
func (a *App) SelectUploadFile() (string, error) {
	filePattern := strings.Join(lo.Map(sydney.BingAllowedFileExtensions, func(item string, index int) string {
		return "*." + item
	}), ";")
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open file to upload",
		Filters: []runtime.FileFilter{{
			DisplayName: "Custom Files (" + filePattern + ")",
			Pattern:     filePattern,
		}},
	})
	if err != nil {
		return "", err
	}
	return file, nil
}
func (a *App) SaveTempFileToUploadFromBase64(ext, rawBase64 string) (string, error) {
	if !lo.Contains(sydney.BingAllowedFileExtensions, ext) {
		return "", errors.New("file extension " + ext + " is not allowed")
	}
	v, err := base64.StdEncoding.DecodeString(rawBase64)
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp("", "*."+ext)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, bytes.NewReader(v))
	if err != nil {
		return "", err
	}
	return filepath.Join(os.TempDir(), f.Name()), nil
}

type UploadSydneyDocumentResult struct {
	Canceled bool   `json:"canceled,omitempty"`
	Text     string `json:"text,omitempty"`
	Ext      string `json:"ext,omitempty"`
}

func (a *App) UploadDocument() (UploadSydneyDocumentResult, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open a document to upload",
		Filters: []runtime.FileFilter{{
			DisplayName: "Document Files (*.pdf; *.pptx; *.docx; *.txt; *.md)",
			Pattern:     "*.pdf;*.pptx;*.docx;*.txt;*.md",
		}},
	})
	if err != nil {
		return UploadSydneyDocumentResult{}, err
	}
	if file == "" {
		return UploadSydneyDocumentResult{Canceled: true}, nil
	}
	ext := filepath.Ext(file)
	var docReader util.DocumentReader
	switch ext {
	case ".pdf":
		docReader = util.PDFDocumentReader{}
	case ".docx":
		docReader = util.DocxDocumentReader{}
	case ".pptx":
		docReader = util.PptxDocumentReader{}
	case ".txt", ".md":
		docReader = util.PlainDocumentReader{}
	default:
		return UploadSydneyDocumentResult{}, errors.New("file type " + ext + " not implemented")
	}
	s, err := docReader.Read(file)
	if err != nil {
		return UploadSydneyDocumentResult{}, err
	}
	text := s
	if !docReader.WillSkipPostprocess() {
		text = strings.ReplaceAll(text, "\r", "")
		text = regexp.MustCompile("(?m)^\r+").ReplaceAllString(text, "")
		text = regexp.MustCompile("\n+").ReplaceAllString(text, "\n")
		v, err := json.Marshal(&text)
		if err != nil {
			return UploadSydneyDocumentResult{}, err
		}
		text = string(v)
	}
	return UploadSydneyDocumentResult{
		Text: text,
		Ext:  ext,
	}, nil
}

func (a *App) FetchWebpage(url string) (string, error) {
	_, client, err := util.MakeHTTPClient(a.settings.config.Proxy, 15*time.Second)
	if err != nil {
		return "", err
	}
	resp, err := client.R().Get("https://r.jina.ai/" + url)
	if err != nil {
		return "", err
	}
	if resp.IsErrorState() {
		return "", errors.New("error fetching url: " + resp.GetStatus() + ": " + resp.String())
	}
	return resp.String(), nil
}

func (a *App) GetUser() (string, error) {
	sydneyIns, err := a.createSydney()
	if err != nil {
		return "", err
	}
	return sydneyIns.GetUser()
}

type CheckUpdateResult struct {
	NeedUpdate     bool   `json:"need_update"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	ReleaseURL     string `json:"release_url"`
	ReleaseNote    string `json:"release_note"`
}

func (a *App) CheckUpdate() (CheckUpdateResult, error) {
	empty := CheckUpdateResult{}
	_, client, err := util.MakeHTTPClient(a.settings.config.Proxy, 15*time.Second)
	if err != nil {
		return empty, err
	}
	resp, err := client.R().Get("https://api.github.com/repos/juzeon/SydneyQt/releases")
	if err != nil {
		return empty, err
	}
	var githubRelease []GithubReleaseResponse
	err = json.Unmarshal(resp.Bytes(), &githubRelease)
	if err != nil {
		return empty, err
	}
	if len(githubRelease) == 0 {
		return empty, errors.New("no release found")
	}
	currentVersion, err := goversion.NewVersion(strings.TrimSpace(version))
	if err != nil {
		return empty, err
	}
	latestVersionStr := githubRelease[0].TagName
	latestVersionStr = strings.TrimPrefix(latestVersionStr, "v")
	latestVersion, err := goversion.NewVersion(latestVersionStr)
	if err != nil {
		return empty, err
	}
	needUpdate := false
	if latestVersion.GreaterThan(currentVersion) {
		needUpdate = true
	}
	return CheckUpdateResult{
		NeedUpdate:     needUpdate,
		CurrentVersion: currentVersion.String(),
		LatestVersion:  latestVersion.String(),
		ReleaseURL:     githubRelease[0].HtmlUrl,
		ReleaseNote:    githubRelease[0].Body,
	}, nil
}

func (a *App) GenerateImage(generativeImage sydney.GenerativeImage) (sydney.GenerateImageResult, error) {
	empty := sydney.GenerateImageResult{}
	syd, err := a.createSydney()
	if err != nil {
		return empty, err
	}
	return syd.GenerateImage(generativeImage)
}
func (a *App) GenerateMusic(generativeMusic sydney.GenerativeMusic) (sydney.GenerateMusicResult, error) {
	var empty sydney.GenerateMusicResult
	syd, err := a.createSydney()
	if err != nil {
		return empty, err
	}
	return syd.GenerateMusic(generativeMusic)
}
func (a *App) SaveRemoteJPEGImage(url string) error {
	if strings.Contains(url, "?") {
		url = strings.Split(url, "?")[0]
	}
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "Choose a destination to save the image",
		Filters: []runtime.FileFilter{{
			DisplayName: "JPEG Image Files (*.jpg, *.jpeg)",
			Pattern:     "*.jpg;*.jpeg",
		}},
		DefaultFilename:      "image.jpg",
		CanCreateDirectories: true,
	})
	if err != nil {
		return err
	}
	if filePath == "" { // cancelled
		return nil
	}
	_, client, err := util.MakeHTTPClient(a.settings.config.Proxy, 30*time.Second)
	if err != nil {
		return err
	}
	resp, err := client.R().Get(url)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(filePath, ".jpg") && !strings.HasSuffix(filePath, ".jpeg") {
		filePath += ".jpg"
	}
	return os.WriteFile(filePath, resp.Bytes(), 0644)
}
func (a *App) SaveRemoteFile(extWithoutDot, defaultFilenameWithoutExt, url string) error {
	fn, err := filenamify.FilenamifyV2(
		lo.Ternary(defaultFilenameWithoutExt != "", defaultFilenameWithoutExt, "file") +
			"." + extWithoutDot)
	if err != nil {
		return err
	}
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "Choose a destination to save the file",
		Filters: []runtime.FileFilter{{
			DisplayName: "*." + extWithoutDot,
			Pattern:     "*." + extWithoutDot,
		}},
		DefaultFilename:      fn,
		CanCreateDirectories: true,
	})
	if err != nil {
		return err
	}
	if filePath == "" { // cancelled
		return nil
	}
	_, client, err := util.MakeHTTPClient(a.settings.config.Proxy, 60*time.Second)
	if err != nil {
		return err
	}
	resp, err := client.R().Get(url)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(filePath, "."+extWithoutDot) {
		filePath += "." + extWithoutDot
	}
	return os.WriteFile(filePath, resp.Bytes(), 0644)
}
func (a *App) ExportWorkspace(id int) error {
	workspace, ok := lo.Find(a.settings.config.Workspaces, func(item Workspace) bool {
		return item.ID == id
	})
	if !ok {
		return errors.New("workspace not exist by id: " + strconv.Itoa(id))
	}
	fn, err := filenamify.FilenamifyV2(workspace.Title + ".md")
	if err != nil {
		return err
	}
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "Choose a destination to save the chat",
		Filters: []runtime.FileFilter{{
			DisplayName: "MarkDown Files (*.md)",
			Pattern:     "*.md",
		}},
		CanCreateDirectories: true,
		DefaultFilename:      fn,
	})
	if err != nil {
		return err
	}
	if filePath == "" {
		return nil
	}
	if !strings.HasSuffix(filePath, ".md") {
		filePath += ".md"
	}
	messages := util.GetChatMessage(workspace.Context)
	var out bytes.Buffer
	for _, msg := range messages {
		out.WriteString(fmt.Sprintf("# \\[%s\\](#%s)\n%s\n\n", msg.Role, msg.Type, msg.Content))
	}
	input := strings.TrimSpace(workspace.Input)
	if input != "" {
		out.WriteString("# \\[user\\](#message)\n" + workspace.Input + "\n\n")
	}
	return os.WriteFile(filePath, out.Bytes(), 0644)
}

type ShareGPTRequest struct {
	Title string         `json:"title"`
	Items []ShareGPTItem `json:"items"`
}
type ShareGPTItem struct {
	From  string `json:"from"`
	Value string `json:"value"`
}
type ShareGPTResponse struct {
	Message string `json:"message"` // on error
	ID      string `json:"id"`      // on success
}

func (a *App) ShareWorkspace(id int) error {
	workspace, ok := lo.Find(a.settings.config.Workspaces, func(item Workspace) bool {
		return item.ID == id
	})
	if !ok {
		return errors.New("workspace not exist by id: " + strconv.Itoa(id))
	}
	_, client, err := util.MakeHTTPClient(a.settings.config.Proxy, 5*time.Second)
	if err != nil {
		return err
	}
	resp, err := client.R().SetBody(ShareGPTRequest{
		Title: workspace.Title,
		Items: lo.Map(util.GetChatMessage(workspace.Context), func(item util.ChatMessage, index int) ShareGPTItem {
			from := "gpt"
			if item.Role == "user" {
				from = "human"
			}
			return ShareGPTItem{
				From:  from,
				Value: fmt.Sprintf("[%s](#%s)\n%s", item.Role, item.Type, item.Content),
			}
		}),
	}).Post("https://sharegpt.com/api/conversations")
	if err != nil {
		return err
	}
	if resp.IsErrorState() {
		return errors.New("error status code: " + resp.GetStatus())
	}
	var response ShareGPTResponse
	err = json.Unmarshal(resp.Bytes(), &response)
	if err != nil {
		return err
	}
	if response.Message != "" {
		return errors.New("error posting conversation: " + response.Message)
	}
	err = util.OpenURL("https://sharegpt.com/c/" + response.ID)
	return nil
}

type YoutubeVideoResult struct {
	Details  YoutubeVideoDetails    `json:"details"`
	Captions []util.YtCustomCaption `json:"captions"`
}
type YoutubeVideoDetails struct {
	Title         string   `json:"title"`
	LengthSeconds string   `json:"length_seconds"`
	Description   string   `json:"description"`
	Keywords      []string `json:"keywords"`
	PicURL        string   `json:"pic_url"`
	Author        string   `json:"author"`
}

func (a *App) GetYoutubeVideo(url string) (YoutubeVideoResult, error) {
	var result YoutubeVideoResult
	yt, err := util.NewYoutube(url, a.settings.config.Proxy)
	if err != nil {
		return result, err
	}
	vd, err := yt.GetVideoDetails()
	if err != nil {
		return result, err
	}
	cp, err := yt.GetCaptions()
	if err != nil {
		return result, err
	}
	th, _ := lo.Last(vd.Thumbnail.Thumbnails)
	result = YoutubeVideoResult{
		Details: YoutubeVideoDetails{
			Title:         vd.Title,
			LengthSeconds: vd.LengthSeconds,
			Description:   vd.ShortDescription,
			Keywords:      vd.Keywords,
			PicURL:        th.Url,
			Author:        vd.Author,
		},
		Captions: cp,
	}
	return result, nil
}

func (a *App) GetYoutubeTranscript(caption util.YtCustomCaption) ([]util.YtTranscriptText, error) {
	return caption.GetTranscript(a.settings.config.Proxy)
}
