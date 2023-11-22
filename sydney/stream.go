package sydney

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"log"
	"net/url"
	"nhooyr.io/websocket"
	"strings"
	"sydneyqt/util"
	"time"
)

func (o *Sydney) AskStream(options AskStreamOptions) <-chan Message {
	out := make(chan Message)
	ch := o.AskStreamRaw(options)
	go func() {
		defer close(out)
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
		for msg := range ch {
			if msg.Error != nil {
				log.Println("error: " + msg.Error.Error())
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
						Text: messageHiddenText,
					}
				case "InternalSearchResult":
					var links []string
					if strings.Contains(messageHiddenText,
						"Web search returned no relevant result") {
						out <- Message{
							Type: MessageTypeSearchResult,
							Text: messageHiddenText,
						}
						continue
					}
					if !gjson.Valid(messageText) {
						log.Println("Error when parsing InternalSearchResult: " + messageText)
						continue
					}
					arr := gjson.Parse(messageText).Array()
					for _, group := range arr {
						srIndex := 1
						for _, subGroup := range group.Array() {
							links = append(links, fmt.Sprintf("[^%d^][%s](%s)",
								srIndex, subGroup.Get("title").String(), subGroup.Get("url").String()))
							srIndex++
						}
					}
					out <- Message{
						Type: MessageTypeSearchResult,
						Text: strings.Join(links, "\n\n"),
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
					out <- Message{
						Type: MessageTypeGenerativeImage,
						Text: messageText,
					}
				case "":
					if data.Get("arguments.0.cursor").Exists() {
						wrote = 0
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
						}
						sendSuggestedResponses(message)
					}
				default:
					log.Println("Unsupported message type: " + msgType.String())
					log.Println("Triggered by " + options.Prompt + ", response: " + message.Raw)
				}
			} else if data.Get("type").Int() == 2 && data.Get("item.messages").Exists() {
				message := data.Get("item.messages|@reverse|0")
				sendSuggestedResponses(message)
			}
		}
	}()
	return out
}
func (o *Sydney) AskStreamRaw(options AskStreamOptions) <-chan RawMessage {
	msgChan := make(chan RawMessage)
	go func() {
		defer close(msgChan)
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
		headers := util.CopyMap(o.headers)
		headers["Cookie"] = util.FormatCookieString(o.cookies)
		httpHeaders := map[string][]string{}
		for k, v := range headers {
			httpHeaders[k] = []string{v}
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
		if o.noSearch {
			options.Prompt += " #no_search"
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
						MessageType:   []string{"Chat", "SearchQuery"}[util.RandIntInclusive(0, 1)],
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
	}()
	return msgChan
}
