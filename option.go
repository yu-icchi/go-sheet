package sheet

import (
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

type option struct {
	// title タイトル
	title string
	// isDatetime int64形式の数字をDatetime形式(2006-01-02 15:04:05)に変換するか否か
	isDatetime bool
}

func (o *option) reset() {
	o.title = ""
	o.isDatetime = false
}

var optionPool = sync.Pool{
	New: func() interface{} {
		return &option{}
	},
}

func newOption(tag string) *option {
	tags := strings.Split(tag, ",")
	opt := optionPool.Get().(*option)
	for _, tag := range tags {
		if tag == "datetime" {
			opt.isDatetime = true
		}
		if strings.HasPrefix(tag, "title=") {
			tmp := strings.Split(tag, "=")
			if len(tmp) > 1 {
				opt.title = tmp[1]
			}
		}
	}
	return opt
}

func resetOption(opt *option) {
	if opt != nil {
		opt.reset()
		optionPool.Put(opt)
	}
}

func encodeDatetime(v reflect.Value) (interface{}, error) {
	switch n := v.Interface().(type) {
	case time.Time:
		if n.IsZero() {
			return nil, nil
		}
		return n.Format(timeFormat), nil
	case int64:
		if n > 0 {
			return time.Unix(n, 0).Format(timeFormat), nil
		}
		return nil, nil
	}
	return v.Interface(), nil
}

func decodeDatetime(v string, opt *option) (time.Time, error) {
	if opt != nil && opt.isDatetime {
		return time.ParseInLocation(timeFormat, v, time.Local)
	}
	now := time.Now()
	if err := now.UnmarshalText([]byte(v)); err != nil {
		return now, err
	}
	return now, nil
}
