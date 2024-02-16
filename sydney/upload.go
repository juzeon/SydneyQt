package sydney

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/imroc/req/v3"
	"io"
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
		return "", err
	}
	resp, err := client.R().SetFileUpload(req.FileUpload{
		ParamName: "knowledgeRequest",
		GetFileContent: func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(payload)), nil
		},
		ContentType: "application/json",
	}, req.FileUpload{
		ParamName: "imageBase64",
		GetFileContent: func() (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(imageBase64)), nil
		},
		ContentType: "application/octet-stream",
	}).Post("https://www.bing.com/images/kblob")
	var result UploadImageResponse
	err = json.Unmarshal(resp.Bytes(), &result)
	if err != nil {
		return "", err
	}
	if result.BlobId == "" {
		return "", errors.New("blobId is empty")
	}
	return "https://www.bing.com/images/blob?bcid=" + result.BlobId, nil
}
