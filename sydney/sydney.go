package sydney

import (
	"log/slog"
	"strconv"
	"sydneyqt/util"

	"github.com/google/uuid"
	clone "github.com/huandu/go-clone/generic"
)

type Sydney struct {
	debug                 bool
	proxy                 string
	conversationStyle     string
	locale                string
	wssURL                string
	createConversationURL string
	noSearch              bool

	optionsSetMap             map[string][]string
	sliceIDs                  []string
	locationHints             map[string][]LocationHint
	allowedMessageTypes       []string
	headers                   map[string]string
	headersCreateConversation map[string]string
	headersCreateImage        map[string]string
}

func NewSydney(options Options) *Sydney {
	debugOptions := clone.Clone(options)
	debugOptions.Cookies = nil
	slog.Info("New Sydney", "v", debugOptions)
	basicOptionsSet := []string{
		"fluxsydney",
		"iyxapbing",
		"iycapbing",
		"clgalileoall",
		"gencontentv3",
		"nojbf",
	}
	if options.GPT4Turbo {
		basicOptionsSet = append(basicOptionsSet, "dlgpt4t")
	}
	uuidObj, err := uuid.NewUUID()
	if err != nil {
		util.GracefulPanic(err)
	}
	forwardedIP := "1.0.0." + strconv.Itoa(util.RandIntInclusive(1, 255))
	cookies := util.Ternary(options.Cookies == nil, map[string]string{}, options.Cookies)
	return &Sydney{
		debug:             options.Debug,
		proxy:             options.Proxy,
		conversationStyle: util.Ternary(options.ConversationStyle == "", "Creative", options.ConversationStyle),
		locale:            util.Ternary(options.Locale == "", "en-US", options.Locale),
		wssURL: util.Ternary(options.WssDomain == "", "wss://sydney.bing.com/sydney/ChatHub",
			"wss://"+options.WssDomain+"/sydney/ChatHub"),
		noSearch: options.NoSearch,
		createConversationURL: util.Ternary(options.CreateConversationURL == "",
			"https://edgeservices.bing.com/edgesvc/turing/conversation/create", options.CreateConversationURL),
		optionsSetMap: map[string][]string{
			"Creative": append(basicOptionsSet, "h3imaginative"),
			"Balanced": append(basicOptionsSet, "galileo"),
			"Precise":  append(basicOptionsSet, "h3precise"),
		},
		sliceIDs: []string{},
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
			"sec-ch-ua":                   `"Microsoft Edge";v="113", "Chromium";v="113", "Not-A.Brand";v="24"`,
			"sec-ch-ua-arch":              `"x86"`,
			"sec-ch-ua-bitness":           `"64"`,
			"sec-ch-ua-full-version":      `"113.0.1774.50"`,
			"sec-ch-ua-full-version-list": `"Microsoft Edge";v="113.0.1774.50", "Chromium";v="113.0.5672.127", "Not-A.Brand";v="24.0.0.0"`,
			"sec-ch-ua-mobile":            "?0",
			"sec-ch-ua-model":             `""`,
			"sec-ch-ua-platform":          `"Windows"`,
			"sec-ch-ua-platform-version":  `"15.0.0"`,
			"sec-fetch-dest":              "empty",
			"sec-fetch-mode":              "cors",
			"sec-fetch-site":              "same-origin",
			"sec-ms-gec":                  util.GenerateSecMSGec(),
			"sec-ms-gec-version":          "1-115.0.1866.1",
			"x-ms-client-request-id":      uuidObj.String(),
			"x-ms-useragent":              "azsdk-js-api-client-factory/1.0.0-beta.1 core-rest-pipeline/1.10.0 OS/Win32",
			"user-agent":                  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36 Edg/113.0.1774.50",
			"Referer":                     "https://www.bing.com/search?q=Bing+AI&showconv=1",
			"Referrer-Policy":             "origin-when-cross-origin",
			"x-forwarded-for":             forwardedIP,
			"Cookie":                      util.FormatCookieString(cookies),
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
			"Cookie":                      util.FormatCookieString(cookies),
		},
		headersCreateImage: map[string]string{
			"authority":                 "www.bing.com",
			"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"accept-language":           "en-US,en;q=0.9",
			"cache-control":             "no-cache",
			"referer":                   "https://www.bing.com/search?q=Bing+AI&showconv=1",
			"upgrade-insecure-requests": "1",
			"user-agent":                "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36 Edg/112.0.1722.46",
			"x-forwarded-for":           forwardedIP,
			"Sec-Fetch-Dest":            "iframe",
			"Cookie":                    util.FormatCookieString(cookies),
		},
	}
}
