package sydney

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sydneyqt/sydney/internal/hex"
	"sydneyqt/util"
	"time"
)

type GenerateMusicRawResponse struct {
	RawResponse string `json:"RawResponse"`
}
type GenerateMusicRealResponse struct {
	Id               string  `json:"id"`
	Status           string  `json:"status"`
	ErrorMessage     string  `json:"errorMessage"`
	GptPrompt        string  `json:"gptPrompt"`
	Lyrics           string  `json:"lyrics"`
	AudioKey         string  `json:"audioKey"`
	ImageKey         string  `json:"imageKey"`
	VideoKey         string  `json:"videoKey"`
	Duration         float64 `json:"duration"`
	SunoJsonResponse string  `json:"sunoJsonResponse"`
	MusicalStyle     string  `json:"musicalStyle"`
	BingShareHash    string  `json:"bingShareHash"`
}

func (o *Sydney) GenerateMusic(generativeMusic GenerativeMusic) (GenerateMusicResult, error) {
	start := time.Now()
	var empty GenerateMusicResult
	_, client, err := util.MakeHTTPClient(o.proxy, 15*time.Second)
	if err != nil {
		return empty, err
	}
	client.SetCommonHeader("Referer", "https://www.bing.com/search?q=Bing+AI&showconv=1&wlexpsignin=1").
		SetCommonHeader("Cookie", util.FormatCookieString(o.cookies))
	u0 := "https://www.bing.com/videos/music?vdpp=suno&kseed=8000&SFX=3&q=&" +
		"iframeid=" + generativeMusic.IFrameID + "&requestid=" + generativeMusic.RequestID
	resp, err := client.R().Get(u0)
	if err != nil {
		return empty, err
	}
	if resp.IsErrorState() {
		return empty, errors.New("videos/music status: " + resp.GetStatus())
	}
	arr := regexp.MustCompile("skey=(.*?)&amp;").FindStringSubmatch(resp.String())
	if len(arr) < 2 {
		return empty, errors.New("cannot find music creation skey")
	}
	u1 := "https://www.bing.com/videos/api/custom/music?skey=" + arr[1] +
		"&safesearch=Moderate&vdpp=suno&" +
		"requestid=" + generativeMusic.RequestID + "&" +
		"ig=" + hex.NewUpperHex(32) + "&iid=vsn&sfx=1"
	slog.Info("Result URL", "v", u1)
	for i := 0; i < 15; i++ {
		time.Sleep(3 * time.Second)
		resp, err = client.R().SetHeader("Referer", u0).Get(u1)
		if err != nil {
			return empty, err
		}
		var rawResp GenerateMusicRawResponse
		err = json.Unmarshal(resp.Bytes(), &rawResp)
		if err != nil {
			return empty, fmt.Errorf("cannot unmarshal raw music response: %w", err)
		}
		var realResp GenerateMusicRealResponse
		err = json.Unmarshal([]byte(rawResp.RawResponse), &realResp)
		if err != nil {
			return empty, fmt.Errorf("cannot unmarshal real music response: %w", err)
		}
		if realResp.Status == "running" {
			slog.Info("Music creation is running")
			continue
		}
		if realResp.Status != "complete" {
			slog.Warn("Music creation failed", "v", realResp)
			return empty, errors.New("music creation failed: " + realResp.ErrorMessage)
		}
		return GenerateMusicResult{
			GenerativeMusic: generativeMusic,
			CoverImgURL:     "https://th.bing.com/th?&id=" + realResp.ImageKey,
			AudioURL:        "https://th.bing.com/th?&id=" + realResp.AudioKey,
			VideoURL:        "https://th.bing.com/th?&id=" + realResp.VideoKey,
			MusicDuration:   time.Duration(realResp.Duration * float64(time.Second)),
			MusicalStyle:    realResp.MusicalStyle,
			Title:           realResp.GptPrompt,
			Lyrics:          realResp.Lyrics,
			TimeElapsed:     time.Since(start),
		}, nil
	}
	return empty, errors.New("music creation timeout")
}
