package sydney

import (
	"github.com/samber/lo"
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
	bypassServer          string

	optionsSet          []string
	sliceIDs            []string
	locationHints       map[string][]LocationHint
	allowedMessageTypes []string
	headers             func() map[string]string
	cookies             map[string]string
}

func NewSydney(options Options) *Sydney {
	debugOptions := clone.Clone(options)
	debugOptions.Cookies = nil
	slog.Info("New Sydney", "v", debugOptions)

	uuidObj, err := uuid.NewUUID()
	if err != nil {
		util.GracefulPanic(err)
	}
	optionsSet := []string{
		"fluxcopilot",
		// no jailbreak filter
		"nojbf",
		"iyxapbing",
		"iycapbing",
		"dgencontentv3",
		"nointernalsugg",
		"disable_telemetry",
		"machine_affinity",
		"streamf",
		// code interpreter
		"codeint",
		"langdtwb",
		"fdwtlst",
		"fluxprod",
		"eredirecturl",
		// may related to image search
		"gptvnodesc",
		"gptvnoex",
	}
	forwardedIP := "1.0.0." + strconv.Itoa(util.RandIntInclusive(1, 255))
	cookies := util.Ternary(options.Cookies == nil, map[string]string{}, options.Cookies)
	options.ConversationStyle = lo.Ternary(options.ConversationStyle == "",
		"Creative", options.ConversationStyle)
	if options.ConversationStyle == "Creative" && options.UseClassic {
		options.ConversationStyle = "CreativeClassic"
	}
	switch options.ConversationStyle {
	case "Balanced":
		optionsSet = append(optionsSet, "galileo", "gldcl1p")
	case "Precise":
		optionsSet = append(optionsSet, "h3precise")
	case "Creative", "CreativeClassic":
		optionsSet = append(optionsSet)
	}
	if options.NoSearch {
		optionsSet = append(optionsSet, "nosearchall")
	}
	if debugOptionSets := util.ReadDebugOptionSets(); len(debugOptionSets) != 0 {
		optionsSet = debugOptionSets
	}
	slog.Info("Final conversation options", "options", optionsSet, "tone", options.ConversationStyle)
	return &Sydney{
		debug:             options.Debug,
		proxy:             options.Proxy,
		conversationStyle: options.ConversationStyle,
		locale:            util.Ternary(options.Locale == "", "en-US", options.Locale),
		wssURL: util.Ternary(options.WssDomain == "", "wss://sydney.bing.com/sydney/ChatHub",
			"wss://"+options.WssDomain+"/sydney/ChatHub"),
		createConversationURL: util.Ternary(options.CreateConversationURL == "",
			"https://edgeservices.bing.com/edgesvc/turing/conversation/create", options.CreateConversationURL),
		bypassServer: options.BypassServer,
		optionsSet:   optionsSet,
		sliceIDs:     []string{},
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
			"InternalLoaderMessage",
			"Progress",
			"GenerateContentQuery",
			"SearchQuery",
			"GeneratedCode",
		},
		headers: func() map[string]string {
			return map[string]string{
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
			}
		},
		cookies: cookies,
	}
}
