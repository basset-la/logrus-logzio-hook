package logzio

import (
	"io"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	endpoint = "listener.logz.io:5050"
	protocol = "udp"
)

var defaultFormatter = &logrus.JSONFormatter{}

// Hook represents a logrus hook
type Hook struct {
	conn      io.Writer
	formatter logrus.Formatter
	fields    logrus.Fields
}

// NewHook returns a new instance of Hook
func NewHook(token string, appName string, fields logrus.Fields) (*Hook, error) {

	f := logrus.Fields{}
	f["token"] = token
	f["appname"] = appName

	merge(f, fields)

	conn, err := net.Dial(protocol, endpoint)

	if err != nil {
		return nil, err
	}

	return &Hook{
		conn:      conn,
		formatter: defaultFormatter,
		fields:    f,
	}, nil

}

// SetFormatter lets you override the default formatter
func (h *Hook) SetFormatter(f logrus.Formatter) {
	h.formatter = f
}

// ClearAllFields clears all fields
func (h *Hook) ClearAllFields(f logrus.Fields) {
	merge(h.fields, f)
}

// Levels returns logging levels
func (h *Hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

// Merge f2 into f1 without overriding fields
func merge(f1 logrus.Fields, f2 logrus.Fields) {
	for k, v := range f2 {
		if _, ok := f1[k]; !ok {
			f1[k] = v
		}
	}
}

// Fire send the entry to Logz.io
func (h *Hook) Fire(entry *logrus.Entry) error {

	merge(entry.Data, h.fields)
	data, err := h.formatter.Format(entry)

	if err != nil {
		return err
	}

	// Horrible hacks
	dataStr := string(data)
	dataStr = strings.Replace(dataStr, "\"msg\":", "\"message\":", 1)
	// Logz.io does not accept level text
	dataStr = strings.Replace(dataStr, "\"level\":\"panic\"", "\"level\":0", 1)
	dataStr = strings.Replace(dataStr, "\"level\":\"fatal\"", "\"level\":1", 1)
	dataStr = strings.Replace(dataStr, "\"level\":\"error\"", "\"level\":2", 1)
	dataStr = strings.Replace(dataStr, "\"level\":\"warning\"", "\"level\":3", 1)
	dataStr = strings.Replace(dataStr, "\"level\":\"info\"", "\"level\":4", 1)
	dataStr = strings.Replace(dataStr, "\"level\":\"debug\"", "\"level\":5", 1)
	data = []byte(dataStr)

	_, err = h.conn.Write(data)

	if err != nil {
		return err
	}

	return nil
}
