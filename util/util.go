package util

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/samber/lo"

	"github.com/ncruces/zenity"
	getproxy "github.com/rapid7/go-get-proxied/proxy"
	tlsx "github.com/refraction-networking/utls"
)

func RandIntInclusive(min int, max int) int {
	return min + rand.Intn(max-min+1)
}
func Ternary[T any](expression bool, trueResult T, falseResult T) T {
	if expression {
		return trueResult
	} else {
		return falseResult
	}
}
func MakeHTTPClient(proxy string, timeout time.Duration) (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	c, err := tlsx.UTLSIdToSpec(tlsx.HelloChrome_102)
	if err != nil {
		return nil, err
	}
	transport.DisableKeepAlives = false
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         c.TLSVersMin,
		MaxVersion:         c.TLSVersMax,
		CipherSuites:       c.CipherSuites,
		ClientSessionCache: tls.NewLRUClientSessionCache(32),
	}
	if proxy != "" { // user filled proxy
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	} else { // try to get system proxy
		log.SetOutput(io.Discard) // FIXME this is a dirty fix that may result in concurrency problems
		proxies := []getproxy.Proxy{
			getproxy.NewProvider("").GetHTTPProxy("https://www.bing.com"),
			getproxy.NewProvider("").GetHTTPSProxy("https://www.bing.com"),
			getproxy.NewProvider("").GetSOCKSProxy("https://www.bing.com"),
		}
		log.SetOutput(os.Stdout)
		var sysProxy getproxy.Proxy
		for _, p := range proxies {
			p := p
			if p != nil {
				sysProxy = p
				break
			}
		}
		if sysProxy != nil { // valid system proxy
			transport.Proxy = http.ProxyURL(sysProxy.URL())
		}
	}
	client := &http.Client{}
	client.Transport = transport
	if timeout != time.Duration(0) {
		client.Timeout = timeout
	}
	return client, nil
}
func FormatCookieString(cookies map[string]string) string {
	str := ""
	for k, v := range cookies {
		str += k + "=" + url.PathEscape(v) + "; "
	}
	return str
}
func CopyMap[T comparable, E any](source map[T]E) map[T]E {
	res := map[T]E{}
	for k, v := range source {
		res[k] = v
	}
	return res
}
func CreateTimeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
func CreateCancelContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}
func MustGenerateRandomHex(length int) string {
	randomBytes := make([]byte, length)
	_, err := rand.New(rand.NewSource(time.Now().Unix())).Read(randomBytes)
	if err != nil {
		GracefulPanic(err)
	}
	randomString := hex.EncodeToString(randomBytes)
	return randomString
}

type FileCookie struct {
	Domain         string      `json:"domain"`
	ExpirationDate float64     `json:"expirationDate"`
	HostOnly       bool        `json:"hostOnly"`
	HttpOnly       bool        `json:"httpOnly"`
	Name           string      `json:"name"`
	Path           string      `json:"path"`
	SameSite       string      `json:"sameSite"`
	Secure         bool        `json:"secure"`
	Session        bool        `json:"session"`
	StoreId        interface{} `json:"storeId"`
	Value          string      `json:"value"`
}

func ReadCookiesFile() (map[string]string, error) {
	res := map[string]string{}
	v, err := os.ReadFile("cookies.json")
	if err != nil {
		return res, nil
	}
	var cookies []FileCookie
	err = json.Unmarshal(v, &cookies)
	if err != nil {
		return res, errors.New("failed to json.Unmarshal content of cookie file")
	}
	for _, cookie := range cookies {
		res[cookie.Name] = cookie.Value
	}
	return res, nil
}
func Map[T any, E any](arr []T, function func(value T) E) []E {
	var result []E
	for _, item := range arr {
		result = append(result, function(item))
	}
	return result
}
func FindFirst[T any](arr []T, function func(value T) bool) (T, bool) {
	var empty T
	for _, item := range arr {
		if function(item) {
			return item, true
		}
	}
	return empty, false
}
func ConvertImageToJpg(img []byte) ([]byte, error) {
	// Decode the image from the []byte
	src, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return nil, err
	}
	// Create a buffer to store the converted image
	var buf bytes.Buffer
	// Encode the image as jpg with quality 80
	err = jpeg.Encode(&buf, src, &jpeg.Options{Quality: 80})
	if err != nil {
		return nil, err
	}
	// Return the buffer as a []byte
	return buf.Bytes(), nil
}
func GenerateSecMSGec() string {
	// Create a new local random generator
	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)

	// Create a byte slice of length 32
	randomBytes := make([]byte, 32)

	// Fill the slice with random bytes
	for i := range randomBytes {
		randomBytes[i] = byte(rng.Intn(256))
	}

	// Convert to hexadecimal
	return hex.EncodeToString(randomBytes)
}
func GracefulPanic(err error) {
	_, file, line, _ := runtime.Caller(1)
	zenity.Error(fmt.Sprintf("Error: %v\nDetails: file(%s), line(%d).\n"+
		"Instruction: This is probably an unknown bug. Please take a screenshot and report this issue.",
		err, file, line))
	lo.Must0(OpenURL("https://github.com/juzeon/SydneyQt/issues"))
	os.Exit(-1)
}
func OpenURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
func ReadDebugOptionSets() (debugOptionsSets []string) {
	debugOptionsSetsFile, err := os.ReadFile("debug_options_sets.json")
	if err != nil {
		return
	}
	json.Unmarshal(debugOptionsSetsFile, &debugOptionsSets)
	if len(debugOptionsSets) != 0 {
		slog.Warn("Enable debug options sets", "v", debugOptionsSets)
	}
	return
}
