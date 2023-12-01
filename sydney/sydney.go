package sydney

import (
	"github.com/google/uuid"
	"log/slog"
	"strconv"
	"sydneyqt/util"
)

type Sydney struct {
	debug                 bool
	cookies               map[string]string
	proxy                 string
	conversationStyle     string
	locale                string
	wssURL                string
	createConversationURL string
	noSearch              bool

	basicOptionsSet           []string
	optionsSetMap             map[string][]string
	sliceIDs                  []string
	locationHints             map[string][]LocationHint
	allowedMessageTypes       []string
	headers                   map[string]string
	headersCreateConversation map[string]string
}

func NewSydney(debug bool, cookies map[string]string, proxy string,
	conversationStyle string, locale string, wssDomain string, createConversationURL string, noSearch bool) *Sydney {
	slog.Info("New Sydney", "proxy", proxy, "conversationStyle",
		conversationStyle, "locale", locale, "wssDomain", wssDomain, "noSearch", noSearch)
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
		"nojbf",
	}
	uuidObj, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	forwardedIP := "1.0.0." + strconv.Itoa(util.RandIntInclusive(1, 255))
	return &Sydney{
		debug:             debug,
		cookies:           cookies,
		proxy:             proxy,
		conversationStyle: conversationStyle,
		locale:            locale,
		wssURL: util.Ternary(wssDomain == "", "wss://sydney.bing.com/sydney/ChatHub",
			"wss://"+wssDomain+"/sydney/ChatHub"),
		noSearch: noSearch,
		createConversationURL: util.Ternary(createConversationURL == "",
			"https://edgeservices.bing.com/edgesvc/turing/conversation/create", createConversationURL),
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
