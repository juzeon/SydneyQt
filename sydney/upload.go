package sydney

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/samber/lo"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sydneyqt/util"
	"time"
)

func (o *Sydney) UploadImage(jpgImgData []byte) (string, error) {
	_, client, err := util.MakeHTTPClient(o.proxy, 60*time.Second)
	if err != nil {
		return "", err
	}
	client.SetCommonHeader("Referer", "https://www.bing.com/search?q=Bing+AI&showconv=1&FORM=hpcodx")
	imageBase64 := base64.StdEncoding.EncodeToString(jpgImgData)
	uploadImagePayload := UploadImagePayload{
		ImageInfo: map[string]any{},
		KnowledgeRequest: KnowledgeRequest{
			InvokedSkills:  []string{"ImageById"},
			SubscriptionId: "Bing.Chat.Multimodal",
			InvokedSkillsRequestData: InvokedSkillsRequestData{
				EnableFaceBlur: false,
			},
			ConvoData: ConvoData{
				Convoid:   "",
				Convotone: o.conversationStyle,
			},
		},
	}
	payload, err := json.Marshal(uploadImagePayload)
	if err != nil {
		return "", fmt.Errorf("cannot marshal uploadImagePayload: %w", err)
	}
	resp, err := client.R().EnableForceMultipart().SetFormData(map[string]string{
		"knowledgeRequest": string(payload),
		"imageBase64":      imageBase64,
	}).Post("https://www.bing.com/images/kblob")
	if err != nil {
		return "", fmt.Errorf("cannot fire upload request: %w", err)
	}
	var result UploadImageResponse
	err = json.Unmarshal(resp.Bytes(), &result)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshal upload response: %w", err)
	}
	if result.BlobId == "" {
		return "", errors.New("blobId is empty")
	}
	return "https://www.bing.com/images/blob?bcid=" + result.BlobId, nil
}

func (o *Sydney) uploadFile(uploadFilePath string, conversation CreateConversationResponse) (UploadFileResult, error) {
	var empty UploadFileResult
	_, client, err := util.MakeHTTPClient(o.proxy, 60*time.Second)
	if err != nil {
		return empty, err
	}
	if !lo.Contains(allowedFileExtensions, strings.TrimPrefix(filepath.Ext(uploadFilePath), ".")) {
		return empty, errors.New("file type " + filepath.Ext(uploadFilePath) + " is not allowed")
	}
	f, err := os.ReadFile(uploadFilePath)
	if err != nil {
		return empty, err
	}
	if len(f) >= 1024*1024 { // 1MB
		return empty, errors.New("file to upload must be less than 1MB")
	}
	var response UploadFileResponse
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+conversation.BearerToken).
		SetHeader("Referer", "https://www.bing.com/search?q=Bing+AI&showconv=1").
		SetHeader("Origin", "https://www.bing.com").
		SetFileUpload(req.FileUpload{
			ParamName: "file",
			FileName:  filepath.Base(uploadFilePath),
			GetFileContent: func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(f)), nil
			},
			FileSize:    int64(len(f)),
			ContentType: "application/octet-stream",
		}).SetFormData(map[string]string{
		"conversationId":              conversation.ConversationId,
		"tone":                        o.conversationStyle,
		"userId":                      conversation.ClientId,
		"enableFileUploadLongContext": "true",
	}).SetSuccessResult(&response).Post("https://sydney.bing.com/sydney/UploadFile")
	if err != nil {
		return empty, err
	}
	if resp.IsErrorState() {
		return empty, errors.New("error status code: " + strconv.Itoa(resp.GetStatusCode()))
	}
	if response.Result.Value != "Success" {
		return empty, errors.New("upload returned failed result: " + response.Result.Message)
	}
	realFileType := fileExtensionToFileType(filepath.Ext(uploadFilePath))
	hiddenText := []UploadFileHiddenText{
		{
			FileName:      response.FileName,
			FileType:      realFileType,
			DocId:         response.DocId,
			IsLongContext: response.IsLongContext,
			UserId:        response.UserId,
			IsBCE:         false,
		},
	}
	v, err := json.Marshal(hiddenText)
	if err != nil {
		return empty, err
	}
	hiddenTextString := string(v)
	result := UploadFileResult{
		Valid:          true,
		Response:       response,
		FileHiddenText: hiddenTextString,
		RealFileType:   realFileType,
	}
	slog.Info("Uploaded file", "result", result)
	return result, nil
}

func fileExtensionToFileType(ext string) string {
	ext = strings.TrimPrefix(ext, ".")
	switch ext {
	case "docx", "rtf":
		return "word"
	case "xlsx":
		return "excel"
	case "pptx":
		return "powerpoint"
	case "pdf":
		return "pdf"
	default:
		return "text"
	}
}

var allowedFileExtensions = []string{
	"rtf",
	"txt",
	"py",
	"ipynb",
	"js",
	"jsx",
	"html",
	"css",
	"java",
	"cs",
	"php",
	"c",
	"cpp",
	"cxx",
	"h",
	"hpp",
	"rs",
	"R",
	"Rmd",
	"swift",
	"go",
	"rb",
	"kt",
	"kts",
	"ts",
	"tsx",
	"m",
	"scala",
	"rs",
	"dart",
	"lua",
	"pl",
	"pm",
	"t",
	"sh",
	"bash",
	"zsh",
	"csv",
	"log",
	"ini",
	"config",
	"json",
	"yaml",
	"yml",
	"toml",
	"lua",
	"sql",
	"md",
	"coffee",
	"tex",
	"latex",
	"pdf",
	"docx",
	"xlsx",
	"pptx",
	"html",
	"wav",
}
