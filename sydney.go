package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
	"strconv"
	"strings"
	"time"
)

const delimiter = '\x1e'

type Sydney struct {
	debug             bool
	cookies           map[string]string
	proxy             string
	conversationStyle string
	locale            string
	wssURL            string
	noSearch          bool

	basicOptionsSet           []string
	optionsSetMap             map[string][]string
	sliceIDs                  []string
	locationHints             map[string][]LocationHint
	allowedMessageTypes       []string
	headers                   map[string]string
	headersCreateConversation map[string]string
}
type LocationHint struct {
	Country           string `json:"country"`
	State             string `json:"state"`
	City              string `json:"city"`
	TimezoneOffset    int    `json:"timezoneoffset"`
	CountryConfidence int    `json:"countryConfidence"`
	Center            LatLng `json:"Center"`
	RegionType        int    `json:"RegionType"`
	SourceType        int    `json:"SourceType"`
}
type LatLng struct {
	Latitude  float64 `json:"Latitude"`
	Longitude float64 `json:"Longitude"`
}
type CreateConversationResult struct {
	Value   string `json:"value"`
	Message string `json:"message"`
}
type CreateConversationResponse struct {
	ConversationId        string                   `json:"conversationId"`
	ClientId              string                   `json:"clientId"`
	Result                CreateConversationResult `json:"result"`
	SecAccessToken        string                   `json:"secAccessToken"`
	ConversationSignature string                   `json:"conversationSignature"`
}

func NewSydney(debug bool, cookies map[string]string, proxy string,
	conversationStyle string, locale string, wssDomain string, noSearch bool) *Sydney {
	basicOptionsSet := []string{
		"nlu_direct_response_filter",
		"deepleo",
		"disable_emoji_spoken_text",
		"responsible_ai_policy_235",
		"enablemm",
		"iycapbing",
		"iyxapbing",
		"dv3sugg",
		"iyoloxap",
		"iyoloneutral",
		"gencontentv3",
		"nojbfedge",
	}
	uuidObj, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	forwardedIP := "1.0.0." + strconv.Itoa(RandIntInclusive(1, 255))
	return &Sydney{
		debug:             debug,
		cookies:           cookies,
		proxy:             proxy,
		conversationStyle: conversationStyle,
		locale:            locale,
		wssURL: Ternary(wssDomain == "", "wss://sydney.bing.com/sydney/ChatHub",
			"wss://"+wssDomain+"/sydney/ChatHub"),
		noSearch:        noSearch,
		basicOptionsSet: basicOptionsSet,
		optionsSetMap: map[string][]string{
			"Creative": append(basicOptionsSet, "h3imaginative"),
			"Balanced": append(basicOptionsSet, "galileo"),
			"Precise":  append(basicOptionsSet, "h3precise"),
		},
		sliceIDs: []string{
			"winmuid1tf",
			"newmma-prod",
			"imgchatgptv2",
			"tts2",
			"voicelang2",
			"anssupfotest",
			"emptyoson",
			"tempcacheread",
			"temptacache",
			"ctrlworkpay",
			"winlongmsg2tf",
			"628fabocs0",
			"531rai268s0",
			"602refusal",
			"621alllocs0",
			"621docxfmtho",
			"621preclsvn",
			"330uaug",
			"529rweas0",
			"0626snptrcs0",
			"619dagslnv1nr",
		},
		locationHints: map[string][]LocationHint{
			"zh-CN": {
				{
					Country:           "China",
					State:             "",
					City:              "Beijing",
					TimezoneOffset:    8,
					CountryConfidence: 8,
					Center: LatLng{
						Latitude:  39.9042,
						Longitude: 116.4074,
					},
					RegionType: 2,
					SourceType: 1,
				},
			},
			"en-US": {
				{
					Country:           "United States",
					State:             "California",
					City:              "Los Angeles",
					TimezoneOffset:    8,
					CountryConfidence: 8,
					Center: LatLng{
						Latitude:  34.0536909,
						Longitude: -118.242766,
					},
					RegionType: 2,
					SourceType: 1,
				},
			},
		},
		allowedMessageTypes: []string{
			"ActionRequest",
			"Chat",
			"Context",
			"InternalSearchQuery",
			"InternalSearchResult",
			"Disengaged",
			"InternalLoaderMessage",
			"Progress",
			"RenderCardRequest",
			"AdsQuery",
			"SemanticSerp",
			"GenerateContentQuery",
			"SearchQuery",
		},
		headers: map[string]string{
			"accept":                      "application/json",
			"accept-language":             "en-US,en;q=0.9",
			"content-type":                "application/json",
			"sec-ch-ua":                   "\"Not_A Brand\";v=\"99\", Microsoft Edge\";v=\"110\", \"Chromium\";v=\"110\"",
			"sec-ch-ua-arch":              "\"x86\"",
			"sec-ch-ua-bitness":           "\"64\"",
			"sec-ch-ua-full-version":      "\"109.0.1518.78\"",
			"sec-ch-ua-full-version-list": "\"Chromium\";v=\"110.0.5481.192\", \"Not A(Brand\";v=\"24.0.0.0\", \"Microsoft Edge\";v=\"110.0.1587.69\"",
			"sec-ch-ua-mobile":            "?0",
			"sec-ch-ua-model":             "",
			"sec-ch-ua-platform":          "\"Windows\"",
			"sec-ch-ua-platform-version":  "\"15.0.0\"",
			"sec-fetch-dest":              "empty",
			"sec-fetch-mode":              "cors",
			"sec-fetch-site":              "same-origin",
			"x-ms-client-request-id":      uuidObj.String(),
			"x-ms-useragent":              "azsdk-js-api-client-factory/1.0.0-beta.1 core-rest-pipeline/1.10.0 OS/Win32",
			"Referer":                     "https://www.bing.com/search?q=Bing+AI&showconv=1&FORM=hpcodx",
			"Referrer-Policy":             "origin-when-cross-origin",
			"x-forwarded-for":             forwardedIP,
		},
		headersCreateConversation: map[string]string{
			"authority":                   "www.bing.com",
			"accept":                      "application/json",
			"accept-language":             "en-US,en;q=0.9",
			"cache-control":               "max-age=0",
			"sec-ch-ua":                   `"Chromium";v="110", "Not A(Brand";v="24", "Microsoft Edge";v="110"`,
			"sec-ch-ua-arch":              `"x86"`,
			"sec-ch-ua-bitness":           `"64"`,
			"sec-ch-ua-full-version":      `"110.0.1587.69"`,
			"sec-ch-ua-full-version-list": `"Chromium";v="110.0.5481.192", "Not A(Brand";v="24.0.0.0", "Microsoft Edge";v="110.0.1587.69"`,
			"sec-ch-ua-mobile":            `"?0"`,
			"sec-ch-ua-model":             `""`,
			"sec-ch-ua-platform":          `"Windows"`,
			"sec-ch-ua-platform-version":  `"15.0.0"`,
			"upgrade-insecure-requests":   "1",
			"user-agent":                  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36 Edg/112.0.1722.46",
			"x-edge-shopping-flag":        "1",
			"x-forwarded-for":             forwardedIP,
		},
	}
}

func (o *Sydney) CreateConversation() (CreateConversationResponse, error) {
	client, err := MakeHTTPClient(o.proxy, 10*time.Second)
	if err != nil {
		return CreateConversationResponse{}, err
	}
	req, err := http.NewRequest("GET", "https://edgeservices.bing.com/edgesvc/turing/conversation/create", nil)
	if err != nil {
		return CreateConversationResponse{}, err
	}
	req.Header.Set("Cookie", FormatCookieString(o.cookies))
	for k, v := range o.headersCreateConversation {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return CreateConversationResponse{}, err
	}
	defer resp.Body.Close()
	bodyV, err := io.ReadAll(resp.Body)
	if err != nil {
		return CreateConversationResponse{}, err
	}
	if resp.StatusCode != 200 {
		return CreateConversationResponse{}, errors.New("Authentication failed: " + string(bodyV))
	}
	var response CreateConversationResponse
	err = json.Unmarshal(bodyV, &response)
	if err != nil {
		return CreateConversationResponse{}, err
	}
	if response.Result.Value == "UnauthorizedRequest" {
		return CreateConversationResponse{}, errors.New(response.Result.Message)
	}
	if value := resp.Header.Get("X-Sydney-Encryptedconversationsignature"); value != "" {
		response.SecAccessToken = value
	}
	if o.debug {
		log.Printf("%#v\n", response)
	}
	return response, nil
}

type Message struct {
	Data  string
	Error error
}
type ChatMessage struct {
	Arguments    []Argument `json:"arguments"`
	InvocationId string     `json:"invocationId"`
	Target       string     `json:"target"`
	Type         int        `json:"type"`
}
type Argument struct {
	OptionsSets           []string          `json:"optionsSets"`
	Source                string            `json:"source"`
	AllowedMessageTypes   []string          `json:"allowedMessageTypes"`
	SliceIds              []string          `json:"sliceIds"`
	Verbosity             string            `json:"verbosity"`
	Scenario              string            `json:"scenario"`
	TraceId               string            `json:"traceId"`
	RequestId             string            `json:"requestId"`
	IsStartOfSession      bool              `json:"isStartOfSession"`
	Message               ArgumentMessage   `json:"message"`
	Tone                  string            `json:"tone"`
	ConversationSignature any               `json:"conversationSignature"`
	Participant           Participant       `json:"participant"`
	SpokenTextMode        string            `json:"spokenTextMode"`
	ConversationId        string            `json:"conversationId"`
	PreviousMessages      []PreviousMessage `json:"previousMessages"`
}
type ArgumentMessage struct {
	Locale        string         `json:"locale"`
	Market        string         `json:"market"`
	Region        string         `json:"region"`
	LocationHints []LocationHint `json:"locationHints"`
	Author        string         `json:"author"`
	InputMethod   string         `json:"inputMethod"`
	Text          string         `json:"text"`
	MessageType   string         `json:"messageType"`
	RequestId     string         `json:"requestId"`
	MessageId     string         `json:"messageId"`
	ImageUrl      any            `json:"imageUrl"`
}
type Participant struct {
	Id string `json:"id"`
}
type PreviousMessage struct {
	Author      string `json:"author"`
	Description string `json:"description"`
	ContextType string `json:"contextType"`
	MessageType string `json:"messageType"`
	MessageId   string `json:"messageId"`
}

func (o *Sydney) AskStream(
	stopCtx context.Context,
	conversation CreateConversationResponse,
	prompt string,
	webpageContext string,
	imageURL string,
) <-chan Message {
	msgChan := make(chan Message)
	go func() {
		defer close(msgChan)
		client, err := MakeHTTPClient(o.proxy, 0)
		if err != nil {
			msgChan <- Message{
				Error: err,
			}
			return
		}
		messageID, err := uuid.NewUUID()
		if err != nil {
			msgChan <- Message{
				Error: err,
			}
			return
		}
		headers := CopyMap(o.headers)
		headers["Cookie"] = FormatCookieString(o.cookies)
		httpHeaders := map[string][]string{}
		for k, v := range headers {
			httpHeaders[k] = []string{v}
		}
		ctx, cancel := CreateTimeoutContext(10 * time.Second)
		defer cancel()
		connRaw, resp, err := websocket.Dial(ctx,
			o.wssURL+Ternary(conversation.SecAccessToken != "", "?sec_access_token="+
				url.QueryEscape(conversation.SecAccessToken), ""),
			&websocket.DialOptions{
				HTTPClient: client,
				HTTPHeader: httpHeaders,
			})
		if err != nil {
			msgChan <- Message{
				Error: err,
			}
			return
		}
		if resp.StatusCode != 101 {
			msgChan <- Message{
				Error: errors.New("cannot establish a websocket connection"),
			}
			return
		}
		defer connRaw.CloseNow()
		select {
		case <-stopCtx.Done():
			return
		default:
		}
		connRaw.SetReadLimit(-1)
		conn := &Conn{Conn: connRaw, debug: o.debug}
		err = conn.WriteWithTimeout([]byte(`{"protocol": "json", "version": 1}`))
		if err != nil {
			msgChan <- Message{
				Error: err,
			}
			return
		}
		conn.ReadWithTimeout()
		err = conn.WriteWithTimeout([]byte(`{"type": 6}`))
		if err != nil {
			msgChan <- Message{
				Error: err,
			}
			return
		}
		if o.noSearch {
			prompt += " #no_search"
		}
		chatMessage := ChatMessage{
			Arguments: []Argument{
				{
					OptionsSets:         o.optionsSetMap[o.conversationStyle],
					Source:              "cib",
					AllowedMessageTypes: o.allowedMessageTypes,
					SliceIds:            o.sliceIDs,
					Verbosity:           "verbose",
					Scenario:            "SERP",
					TraceId:             MustGenerateRandomHex(16),
					RequestId:           messageID.String(),
					IsStartOfSession:    true,
					Message: ArgumentMessage{
						Locale:        o.locale,
						Market:        o.locale,
						Region:        o.locale[len(o.locale)-2:],
						LocationHints: o.locationHints[o.locale],
						Author:        "user",
						InputMethod:   "Keyboard",
						Text:          prompt,
						MessageType:   []string{"Chat", "SearchQuery"}[RandIntInclusive(0, 1)],
						RequestId:     messageID.String(),
						MessageId:     messageID.String(),
						ImageUrl:      Ternary[any](imageURL == "", nil, imageURL),
					},
					Tone: o.conversationStyle,
					ConversationSignature: Ternary[any](conversation.ConversationSignature == "",
						nil, conversation.ConversationSignature),
					Participant:    Participant{Id: conversation.ClientId},
					SpokenTextMode: "None",
					ConversationId: conversation.ConversationId,
					PreviousMessages: []PreviousMessage{
						{
							Author:      "user",
							Description: webpageContext,
							ContextType: "WebPage",
							MessageType: "Context",
							MessageId:   "discover-web--page-ping-mriduna-----",
						},
					},
				},
			},
			InvocationId: "0",
			Target:       "chat",
			Type:         4,
		}
		chatMessageV, err := json.Marshal(&chatMessage)
		if err != nil {
			msgChan <- Message{
				Error: err,
			}
			return
		}
		err = conn.WriteWithTimeout(chatMessageV)
		if err != nil {
			msgChan <- Message{
				Error: err,
			}
			return
		}
		for {
			select {
			case <-stopCtx.Done():
				return
			default:
			}
			messages, err := conn.ReadWithTimeout()
			if err != nil {
				msgChan <- Message{
					Error: err,
				}
				return
			}
			if time.Now().Unix()%6 == 0 {
				err = conn.WriteWithTimeout([]byte(`{"type": 6}`))
				if err != nil {
					msgChan <- Message{
						Error: err,
					}
					return
				}
			}
			for _, msg := range messages {
				if msg == "" {
					continue
				}
				if !gjson.Valid(msg) {
					msgChan <- Message{
						Error: errors.New("malformed json"),
					}
					return
				}
				result := gjson.Parse(msg)
				if result.Get("type").Int() == 2 && result.Get("item.result.value").String() != "Success" {
					msgChan <- Message{
						Error: errors.New(result.Get("item.result.value").Raw + ": " +
							result.Get("item.result.message").Raw),
					}
					return
				}
				msgChan <- Message{
					Data: msg,
				}
				if result.Get("type").Int() == 2 {
					// finish the conversation
					return
				}
			}
		}
	}()
	return msgChan
}

type Conn struct {
	debug bool
	*websocket.Conn
}

func (o *Conn) WriteWithTimeout(v []byte) error {
	ctx, cancel := CreateTimeoutContext(5 * time.Second)
	defer cancel()
	bytes := append(v, []byte(string(delimiter))...)
	if o.debug {
		log.Println("sending: " + string(bytes))
	}
	return o.Write(ctx, websocket.MessageText, bytes)
}
func (o *Conn) ReadWithTimeout() ([]string, error) {
	ctx, cancel := CreateTimeoutContext(30 * time.Second)
	defer cancel()
	typ, v, err := o.Read(ctx)
	if err != nil {
		return nil, err
	}
	if typ != websocket.MessageText {
		return nil, nil
	}
	if len(v) == 0 {
		return nil, errors.New("no response from server")
	}
	str := string(v)
	arr := strings.Split(str, string(delimiter))
	if o.debug {
		for _, item := range arr {
			log.Println("receiving: " + item)
		}
	}
	return arr, nil
}
