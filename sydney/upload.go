package sydney

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"strings"
	"sydneyqt/util"
	"time"
)

func (o *Sydney) UploadImage(data []byte) (string, error) {
	httpClient, err := util.MakeHTTPClient(o.proxy, 0)
	if err != nil {
		return "", err
	}
	client := resty.New().
		SetTransport(httpClient.Transport).
		SetTimeout(60*time.Second).
		SetHeader("Referer", "https://www.bing.com/search?q=Bing+AI&showconv=1&FORM=hpcodx")
	jpgData, err := util.ConvertImageToJpg(data)
	if err != nil {
		return "", err
	}
	imageBase64 := base64.StdEncoding.EncodeToString(jpgData)
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
	resp, err := client.R().SetMultipartFields(&resty.MultipartField{
		Param:       "knowledgeRequest",
		ContentType: "application/json",
		Reader:      bytes.NewReader(payload),
	}, &resty.MultipartField{
		Param:       "imageBase64",
		ContentType: "application/octet-stream",
		Reader:      strings.NewReader(imageBase64),
	}).Post("https://www.bing.com/images/kblob")
	var result UploadImageResponse
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return "", err
	}
	if result.BlobId == "" {
		return "", errors.New("blobId is empty")
	}
	return "https://www.bing.com/images/blob?bcid=" + result.BlobId, nil
}
