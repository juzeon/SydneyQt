package sydney

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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
