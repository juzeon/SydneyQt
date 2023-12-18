package sydney

import (
	"errors"
	"log/slog"
	"nhooyr.io/websocket"
	"strings"
	"sydneyqt/util"
	"time"
)

type Conn struct {
	debug bool
	*websocket.Conn
}

func (o *Conn) WriteWithTimeout(v []byte) error {
	ctx, cancel := util.CreateTimeoutContext(5 * time.Second)
	defer cancel()
	bytes := append(v, []byte(string(delimiter))...)
	slog.Debug("WriteWithTimeout", "v", string(bytes))
	return o.Write(ctx, websocket.MessageText, bytes)
}
func (o *Conn) ReadWithTimeout() ([]string, error) {
	ctx, cancel := util.CreateTimeoutContext(30 * time.Second)
	defer cancel()
	typ, v, err := o.Read(ctx)
	if err != nil {
		var closeErr websocket.CloseError
		if errors.As(err, &closeErr) && closeErr.Code == websocket.StatusNormalClosure {
			err = errors.Join(err, errors.New("please check if the chat context is too long"))
		}
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
	for _, item := range arr {
		slog.Debug("ReadWithTimeout", "v", item)
	}
	return arr, nil
}
