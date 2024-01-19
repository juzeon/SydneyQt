package sydney

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sydneyqt/util"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
	"nhooyr.io/websocket"
)

var debugOptionSets = util.ReadDebugOptionSets()

func (o *Sydney) AskStream(options AskStreamOptions) <-chan Message {
	out := make(chan Message)
	ch := o.AskStreamRaw(options)
	go func(out chan Message, ch <-chan RawMessage) {
		defer func() {
			slog.Info("AskStream is closing out message channel")
			close(out)
		}()
		wrote := 0
		sendSuggestedResponses := func(message gjson.Result) {
			if message.Get("suggestedResponses").Exists() {
				arr := util.Map(message.Get("suggestedResponses").Array(), func(v gjson.Result) string {
					return v.Get("text").String()
				})
				v, _ := json.Marshal(arr)
				out <- Message{
					Type: MessageTypeSuggestedResponses,
					Text: string(v),
				}
			}
		}
		var sourceAttributes []SourceAttribute
		for msg := range ch {
			if msg.Error != nil {
				slog.Error("Ask stream message", "error", msg.Error)
				out <- Message{
					Type:  MessageTypeError,
					Text:  msg.Error.Error(),
					Error: msg.Error,
				}
				return
			}
			data := gjson.Parse(msg.Data)
			if data.Get("type").Int() == 1 && data.Get("arguments.0.messages").Exists() {
				message := data.Get("arguments.0.messages.0")
				msgType := message.Get("messageType")
				messageText := message.Get("text").String()
				messageHiddenText := message.Get("hiddenText").String()
				switch msgType.String() {
				case "InternalSearchQuery":
					out <- Message{
						Type: MessageTypeSearchQuery,
						Text: messageText,
					}
				case "InternalSearchResult":
					if strings.Contains(messageHiddenText,
						"Web search returned no relevant result") {
						slog.Info("Web search returned no relevant result")
						continue
					}
					if !gjson.Valid(messageText) {
						slog.Error("Error when parsing InternalSearchResult", "messageText", messageText)
						continue
					}
					arr := gjson.Parse(messageText).Array()
					for _, group := range arr {
						group.ForEach(func(key, value gjson.Result) bool {
							for _, subGroup := range value.Array() {
								sourceAttributes = append(sourceAttributes, SourceAttribute{
									Link:  subGroup.Get("url").String(),
									Title: subGroup.Get("title").String(),
								})
							}
							return true
						})
					}
				case "InternalLoaderMessage":
					if message.Get("hiddenText").Exists() {
						out <- Message{
							Type: MessageTypeLoading,
							Text: messageHiddenText,
						}
						continue
					}
					if message.Get("text").Exists() {
						out <- Message{
							Type: MessageTypeLoading,
							Text: messageText,
						}
						continue
					}
					out <- Message{
						Type: MessageTypeLoading,
						Text: message.Raw,
					}
				case "GenerateContentQuery":
					if message.Get("contentType").String() != "IMAGE" {
						continue
					}
					generativeImage := GenerativeImage{
						Text: messageText,
						URL: "https://www.bing.com/images/create?" +
							"partner=sydney&re=1&showselective=1&sude=1&kseed=8500&SFX=4" +
							"&q=" + url.QueryEscape(messageText) + "&iframeid=" +
							message.Get("messageId").String(),
					}
					v, err := json.Marshal(&generativeImage)
					if err != nil {
						util.GracefulPanic(err)
					}
					out <- Message{
						Type: MessageTypeGenerativeImage,
						Text: string(v),
					}
				case "":
					if data.Get("arguments.0.cursor").Exists() {
						wrote = 0
						// extract search result from text block
						if text := strings.TrimSuffix(message.Get("adaptiveCards.0.body.0.text").String(),
							messageText); strings.TrimSpace(text) != "" {
							arr := lo.Filter(lo.Map(strings.Split(text, "\n"), func(item string, index int) string {
								return strings.Trim(item, " \"")
							}), func(item string, index int) bool {
								return item != ""
							})
							re := regexp.MustCompile(`\[(\d+)]: (.*)`)
							var resultSources []SourceAttribute
							for _, line := range arr {
								matches := re.FindStringSubmatch(line)
								if len(matches) == 0 {
									continue
								}
								ix := matches[1]
								link := matches[2]
								sourceAttribute, ok := lo.Find(sourceAttributes, func(item SourceAttribute) bool {
									return item.Link == link
								})
								if !ok {
									continue
								}
								sourceAttribute.Index, _ = strconv.Atoi(ix)
								resultSources = append(resultSources, sourceAttribute)
							}
							var resultArr []string
							for _, src := range resultSources {
								v, _ := json.Marshal(&src)
								resultArr = append(resultArr, "  "+string(v))
							}
							out <- Message{
								Type: MessageTypeSearchResult,
								Text: "[\n" + strings.Join(resultArr, ",\n") + "\n]",
							}
						}
					}
					if message.Get("contentOrigin").String() == "Apology" {
						if wrote != 0 {
							out <- Message{
								Type:  MessageTypeError,
								Text:  "Message revoke detected",
								Error: ErrMessageRevoke,
							}
						} else {
							out <- Message{
								Type:  MessageTypeError,
								Text:  "Looks like the user's message has triggered the Bing filter",
								Error: ErrMessageFiltered,
							}
						}
						return
					} else {
						if wrote < len(messageText) {
							out <- Message{
								Type: MessageTypeMessageText,
								Text: messageText[wrote:],
							}
							wrote = len(messageText)
						} else if wrote > len(messageText) { // Bing deletes some already sent text
							wrote = len(messageText)
						}
						sendSuggestedResponses(message)
					}
				case "Progress":
					contentOrigin := message.Get("contentOrigin").String()
					if contentOrigin == "CodeInterpreter" {
						var loadingMessages []string

						message.Get("adaptiveCards.0.body.0.columns.#.items.#.text").
							ForEach(func(key, value gjson.Result) bool {
								value.ForEach(
									func(key, value gjson.Result) bool {
										loadingMessages = append(loadingMessages, value.String())
										return true
									})
								return true
							})

						out <- Message{
							Type: MessageTypeExecutingTask,
							Text: strings.Join(loadingMessages, " "),
						}

						continue
					}

					fallthrough
				default:
					slog.Warn("Unsupported message type",
						"type", msgType.String(), "triggered-by", options.Prompt, "response", message.Raw)
				}
			} else if data.Get("type").Int() == 2 && data.Get("item.messages").Exists() {
				message := data.Get("item.messages|@reverse|0")
				sendSuggestedResponses(message)
			}
		}
	}(out, ch)
	return out
}
func (o *Sydney) AskStreamRaw(options AskStreamOptions) <-chan RawMessage {
	slog.Info("AskStreamRaw called")
	msgChan := make(chan RawMessage)
	go func(msgChan chan RawMessage) {
		defer func(msgChan chan RawMessage) {
			slog.Info("AskStreamRaw is closing raw message channel")
			close(msgChan)
		}(msgChan)
		client, err := util.MakeHTTPClient(o.proxy, 0)
		if err != nil {
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		messageID, err := uuid.NewUUID()
		if err != nil {
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		httpHeaders := http.Header{}
		for k, v := range o.headers {
			httpHeaders.Set(k, v)
		}
		ctx, cancel := util.CreateTimeoutContext(10 * time.Second)
		defer cancel()
		connRaw, resp, err := websocket.Dial(ctx,
			o.wssURL+util.Ternary(options.Conversation.SecAccessToken != "", "?sec_access_token="+
				url.QueryEscape(options.Conversation.SecAccessToken), ""),
			&websocket.DialOptions{
				HTTPClient: client,
				HTTPHeader: httpHeaders,
			})
		if err != nil {
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		if resp.StatusCode != 101 {
			msgChan <- RawMessage{
				Error: errors.New("cannot establish a websocket connection"),
			}
			return
		}
		defer connRaw.CloseNow()
		select {
		case <-options.StopCtx.Done():
			slog.Info("Exit askStream because of received signal from stopCtx")
			return
		default:
		}
		connRaw.SetReadLimit(-1)
		conn := &Conn{Conn: connRaw, debug: o.debug}
		err = conn.WriteWithTimeout([]byte(`{"protocol": "json", "version": 1}`))
		if err != nil {
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		conn.ReadWithTimeout()
		err = conn.WriteWithTimeout([]byte(`{"type": 6}`))
		if err != nil {
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		optionsSets := o.optionsSetMap[o.conversationStyle]
		if o.noSearch {
			optionsSets = append(optionsSets, "nosearchall")
		}
		if len(debugOptionSets) != 0 {
			optionsSets = debugOptionSets
		}
		chatMessage := ChatMessage{
			Arguments: []Argument{
				{
					OptionsSets:         optionsSets,
					Source:              "edge_coauthor_prod",
					AllowedMessageTypes: o.allowedMessageTypes,
					SliceIds:            o.sliceIDs,
					Verbosity:           "verbose",
					Scenario:            "Underside",
					TraceId:             util.MustGenerateRandomHex(16),
					RequestId:           messageID.String(),
					IsStartOfSession:    true,
					Message: ArgumentMessage{
						Locale:        o.locale,
						Market:        o.locale,
						Region:        o.locale[len(o.locale)-2:],
						LocationHints: o.locationHints[o.locale],
						Author:        "user",
						InputMethod:   "Keyboard",
						Text:          options.Prompt,
						MessageType:   []string{"Chat", "SearchQuery", "CurrentWebpageContextRequest"}[util.RandIntInclusive(0, 2)],
						RequestId:     messageID.String(),
						MessageId:     messageID.String(),
						ImageUrl:      util.Ternary[any](options.ImageURL == "", nil, options.ImageURL),
					},
					Tone: o.conversationStyle,
					ConversationSignature: util.Ternary[any](options.Conversation.ConversationSignature == "",
						nil, options.Conversation.ConversationSignature),
					Participant:    Participant{Id: options.Conversation.ClientId},
					SpokenTextMode: "None",
					ConversationId: options.Conversation.ConversationId,
					PreviousMessages: []PreviousMessage{
						{
							Author:      "user",
							Description: options.WebpageContext,
							ContextType: "WebPage",
							MessageType: "Context",
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
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		err = conn.WriteWithTimeout(chatMessageV)
		if err != nil {
			msgChan <- RawMessage{
				Error: err,
			}
			return
		}
		for {
			select {
			case <-options.StopCtx.Done():
				slog.Info("Exit askStream because of received signal from stopCtx")
				return
			default:
			}
			messages, err := conn.ReadWithTimeout()
			if err != nil {
				msgChan <- RawMessage{
					Error: err,
				}
				return
			}
			if time.Now().Unix()%6 == 0 {
				err = conn.WriteWithTimeout([]byte(`{"type": 6}`))
				if err != nil {
					msgChan <- RawMessage{
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
					msgChan <- RawMessage{
						Error: errors.New("malformed json"),
					}
					return
				}
				result := gjson.Parse(msg)
				if result.Get("type").Int() == 2 && result.Get("item.result.value").String() != "Success" {
					msgChan <- RawMessage{
						Error: errors.New(result.Get("item.result.value").Raw + ": " +
							result.Get("item.result.message").Raw),
					}
					return
				}
				msgChan <- RawMessage{
					Data: msg,
				}
				if result.Get("type").Int() == 2 {
					// finish the conversation
					return
				}
			}
		}
	}(msgChan)
	return msgChan
}
