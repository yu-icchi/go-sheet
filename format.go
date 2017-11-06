package sheet

import (
	"reflect"
	"strings"
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

func genOption(tag string) *option {
	tags := strings.Split(tag, ",")
	opt := &option{}
	for _, tag := range tags {
		if tag == "datetime" {
			opt.isDatetime = true
		}
		if strings.HasPrefix(tag, "title") {
			tmp := strings.Split(tag, "=")
			if len(tmp) > 1 {
				opt.title = tmp[1]
			}
		}
	}
	return opt
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

//func decodeDatetime(v string, opt *option) (reflect.Value, error) {
//
//}
