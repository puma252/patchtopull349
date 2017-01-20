package logrus

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const DefaultTimestampFormat = time.RFC3339

// The Formatter interface is used to implement a custom Formatter. It takes an
// `Entry`. It exposes all the fields, including the default ones:
//
// * `entry.Data["msg"]`. The message passed from Info, Warn, Error ..
// * `entry.Data["time"]`. The timestamp.
// * `entry.Data["level"]. The level the entry was logged at.
//
// Any additional fields added with `WithField` or `WithFields` are also in
// `entry.Data`. Format is expected to return an array of bytes which are then
// logged to `logger.Out`.
type Formatter interface {
	Format(*Entry) ([]byte, error)
}

// This is to not silently overwrite `time`, `msg` and `level` fields when
// dumping it. If this code wasn't there doing:
//
//  logrus.WithField("level", 1).Info("hello")
//
// Would just silently drop the user provided level. Instead with this code
// it'll logged as:
//
//  {"level": "info", "fields.level": 1, "msg": "hello", "time": "..."}
//
// It's not exported because it's still using Data in an opinionated way. It's to
// avoid code duplication between the two default formatters.
func prefixFieldClashes(data Fields, showCaller bool, depth int) {
	if t, ok := data["time"]; ok {
		data["fields.time"] = t
	}
	if m, ok := data["msg"]; ok {
		data["fields.msg"] = m
	}
	if l, ok := data["level"]; ok {
		data["fields.level"] = l
	}

	if showCaller {
		if _, ok := data["caller"]; ok {
			data["fields.caller"] = data["caller"]
		}

		data["caller"] = getcaller(depth)
	}
}

func caller(depth int) (str string) {

	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		str = "???: ?"
	} else {
		str = fmt.Sprint(filepath.Base(file), ":", line)
	}
	return
}

func getcaller(depth int) (str string) {

	MaxDepth := 9 //max search depth
	UnKnownFileInfo := "???: ?"

	d := depth - 2 //the min depth of logrus.Entry.log
	funcName := ""
	str = UnKnownFileInfo

	pc, _, _, ok := runtime.Caller(d)

	for {

		if !ok {
			return
		}
		funcName = runtime.FuncForPC(pc).Name()
		funcNames := strings.Split(funcName, "/")

		if funcNames[len(funcNames)-1] == "logrus.Entry.log" {
			str = caller(d + 3)
			return
		}

		d = d + 1
		pc, _, _, ok = runtime.Caller(d)

		if d > MaxDepth {
			fmt.Printf("Failed to get File infomation, execeed the max depth %d\n", MaxDepth)
			break
		}
	}

	return
}
