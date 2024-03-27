package util

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/imroc/req/v3"
	"github.com/samber/lo"
	"regexp"
	"strconv"
	"time"
)

type Youtube struct {
	VideoID  string
	client   *req.Client
	htmlPage string
}

func NewYoutube(url string, proxy string) (*Youtube, error) {
	arr := regexp.MustCompile("v=([^&]+)").FindStringSubmatch(url)
	if len(arr) == 0 {
		return nil, errors.New("invalid youtube video url: " + url)
	}
	_, client, err := MakeHTTPClient(proxy, 15*time.Second)
	if err != nil {
		return nil, err
	}
	resp, err := client.R().Get("https://www.youtube.com/watch?v=" + arr[1])
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, errors.New("cannot fetch youtube url: " + strconv.Itoa(resp.GetStatusCode()))
	}
	return &Youtube{
		VideoID:  arr[1],
		client:   client,
		htmlPage: resp.String(),
	}, nil
}
func (o *Youtube) GetVideoDetails() (YtVideoDetails, error) {
	var result YtVideoDetails
	arr := regexp.MustCompile("\"videoDetails\":(.*?),\"playerConfig\":").FindStringSubmatch(o.htmlPage)
	if len(arr) == 0 {
		return result, errors.New("cannot find video details")
	}
	err := json.Unmarshal([]byte(arr[1]), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}
func (o *Youtube) GetCaptions() (YtCaptions, error) {
	var result YtCaptions
	arr := regexp.MustCompile("\"captions\":(.*?),\"videoDetails\":").FindStringSubmatch(o.htmlPage)
	if len(arr) == 0 {
		return result, errors.New("cannot find captions")
	}
	err := json.Unmarshal([]byte(arr[1]), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}
func (o *Youtube) GetTranscript(baseURL string, targetLang string) ([]YtTranscriptText, error) {
	url := baseURL + lo.Ternary(targetLang != "", "&tlang="+targetLang, "")
	resp, err := o.client.R().Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, errors.New("cannot fetch transcript: " + strconv.Itoa(resp.GetStatusCode()) + "; url: " + url)
	}
	var transcript YtTranscript
	err = xml.Unmarshal(resp.Bytes(), &transcript)
	if err != nil {
		return nil, err
	}
	return transcript.Texts, nil
}

type YtCaptions struct {
	PlayerCaptionsTracklistRenderer YtPlayerCaptionsTracklistRenderer `json:"playerCaptionsTracklistRenderer"`
}
type YtPlayerCaptionsTracklistRenderer struct {
	CaptionTracks          []YtCaptionTrack        `json:"captionTracks"`
	AudioTracks            []YtAudioTrack          `json:"audioTracks"`
	TranslationLanguages   []YtTranslationLanguage `json:"translationLanguages"`
	DefaultAudioTrackIndex int                     `json:"defaultAudioTrackIndex"`
}
type YtTranslationLanguage struct {
	LanguageCode string     `json:"languageCode"`
	LanguageName YtLangName `json:"languageName"`
}
type YtAudioTrack struct {
	CaptionTrackIndices []int `json:"captionTrackIndices"`
}
type YtCaptionTrack struct {
	BaseUrl        string     `json:"baseUrl"`
	Name           YtLangName `json:"name"`
	VssId          string     `json:"vssId"`
	LanguageCode   string     `json:"languageCode"`
	Kind           string     `json:"kind"`
	IsTranslatable bool       `json:"isTranslatable"`
	TrackName      string     `json:"trackName"`
}
type YtLangName struct {
	SimpleText string `json:"simpleText"`
}
type YtVideoDetails struct {
	VideoId                string      `json:"videoId"`
	Title                  string      `json:"title"`
	LengthSeconds          string      `json:"lengthSeconds"`
	Keywords               []string    `json:"keywords"`
	ChannelId              string      `json:"channelId"`
	IsOwnerViewing         bool        `json:"isOwnerViewing"`
	ShortDescription       string      `json:"shortDescription"`
	IsCrawlable            bool        `json:"isCrawlable"`
	Thumbnail              YtThumbnail `json:"thumbnail"`
	AllowRatings           bool        `json:"allowRatings"`
	ViewCount              string      `json:"viewCount"`
	Author                 string      `json:"author"`
	IsLowLatencyLiveStream bool        `json:"isLowLatencyLiveStream"`
	IsPrivate              bool        `json:"isPrivate"`
	IsUnpluggedCorpus      bool        `json:"isUnpluggedCorpus"`
	LatencyClass           string      `json:"latencyClass"`
	IsLiveContent          bool        `json:"isLiveContent"`
}
type YtThumbnail struct {
	Thumbnails []YtThumbNailItem `json:"thumbnails"`
}
type YtThumbNailItem struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type YtTranscript struct {
	XMLName xml.Name           `xml:"transcript"`
	Texts   []YtTranscriptText `xml:"text"`
}
type YtTranscriptText struct {
	Start float64 `xml:"start,attr"`
	Dur   float64 `xml:"dur,attr"`
	Value string  `xml:",chardata"`
}
