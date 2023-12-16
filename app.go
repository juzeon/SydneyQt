package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	goversion "github.com/hashicorp/go-version"
	"github.com/life4/genesis/slices"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkoukk/tiktoken-go"
	"github.com/samber/lo"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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
}

// NewApp creates a new App application struct
func NewApp(settings *Settings) *App {
	return &App{settings: settings}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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

func (a *App) UploadSydneyImage() (UploadSydneyImageResult, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open an image to upload",
		Filters: []runtime.FileFilter{{
			DisplayName: "Image Files (*.jpg; *.jpeg; *.png; *.gif)",
			Pattern:     "*.jpg;*.jpeg;*.png;*.gif",
		}},
	})
	if err != nil {
		return UploadSydneyImageResult{}, err
	}
	if file == "" {
		return UploadSydneyImageResult{Canceled: true}, nil
	}
	sydneyIns, err := a.createSydney()
	if err != nil {
		return UploadSydneyImageResult{}, err
	}
	v, err := os.ReadFile(file)
	if err != nil {
		return UploadSydneyImageResult{}, err
	}
	jpgData, err := util.ConvertImageToJpg(v)
	if err != nil {
		return UploadSydneyImageResult{}, err
	}
	url, err := sydneyIns.UploadImage(jpgData)
	if err != nil {
		return UploadSydneyImageResult{}, err
	}
	return UploadSydneyImageResult{
		Base64URL: "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(jpgData),
		BingURL:   url,
	}, err
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
			DisplayName: "Document Files (*.pdf; *.pptx; *.docx)",
			Pattern:     "*.pdf;*.pptx;*.docx",
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

type FetchWebpageResult struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (a *App) FetchWebpage(url string) (FetchWebpageResult, error) {
	empty := FetchWebpageResult{}
	rawClient, err := util.MakeHTTPClient(a.settings.config.Proxy, 0)
	if err != nil {
		return empty, err
	}
	client := resty.New().SetTransport(rawClient.Transport).SetTimeout(15 * time.Second)
	resp, err := client.R().Get(url)
	if err != nil {
		return empty, err
	}
	content := string(resp.Body())
	title := ""
	if doc, err := goquery.NewDocumentFromReader(strings.NewReader(content)); err == nil {
		title = doc.Find("title").Text()
		doc.Find("script").Remove()
		doc.Find("style").Remove()
		text := bluemonday.StripTagsPolicy().Sanitize(doc.Text())
		text = regexp.MustCompile(" {2,}").ReplaceAllString(text, "  ")
		lines := slices.Filter(strings.Split(text, "\n"), func(s string) bool {
			return strings.TrimSpace(s) != ""
		})
		content = strings.Join(lines, "\n")
	}
	return FetchWebpageResult{
		Title:   title,
		Content: content,
	}, nil
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
	hClient, err := util.MakeHTTPClient(a.settings.config.Proxy, 0)
	if err != nil {
		return empty, err
	}
	client := resty.New().SetTimeout(15 * time.Second).SetTransport(hClient.Transport)
	resp, err := client.R().Get("https://api.github.com/repos/juzeon/SydneyQt/releases")
	if err != nil {
		return empty, err
	}
	var githubRelease []GithubReleaseResponse
	err = json.Unmarshal(resp.Body(), &githubRelease)
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
func (a *App) SaveRemoteJPEGImage(url string) error {
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "Choose a destination to save the image",
		Filters: []runtime.FileFilter{{
			DisplayName: "JPEG Image Files (*.jpg, *.jpeg)",
			Pattern:     "*.jpg;*.jpeg",
		}},
		CanCreateDirectories: true,
	})
	if err != nil {
		return err
	}
	if filePath == "" { // cancelled
		return nil
	}
	hClient, err := util.MakeHTTPClient(a.settings.config.Proxy, 0)
	if err != nil {
		return err
	}
	client := resty.New().SetTimeout(30 * time.Second).SetTransport(hClient.Transport)
	resp, err := client.R().Get(url)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(filePath, ".jpg") && !strings.HasSuffix(filePath, ".jpeg") {
		filePath += ".jpg"
	}
	return os.WriteFile(filePath, resp.Body(), 0644)
}
func (a *App) ExportWorkspace(id int) error {
	workspace, ok := lo.Find(a.settings.config.Workspaces, func(item Workspace) bool {
		return item.ID == id
	})
	if !ok {
		return errors.New("workspace not exist by id: " + strconv.Itoa(id))
	}
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "Choose a destination to save the chat",
		Filters: []runtime.FileFilter{{
			DisplayName: "MarkDown Files (*.md)",
			Pattern:     "*.md",
		}},
		CanCreateDirectories: true,
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
