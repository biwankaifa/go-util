package time

import (
	"github.com/golang-module/carbon"
	"strconv"
	"time"
)

func DiffForHumans(time time.Time) string {
	hTime := carbon.Time2Carbon(time)
	nTime := carbon.Now(carbon.Local)

	t := nTime.ToTimestamp() - hTime.ToTimestamp()

	f := []struct {
		value int
		s     string
	}{
		{3600, "小时"},
		{60, "分钟"},
	}

	if t <= 300 {
		return "刚刚"
	} else if t <= 3600*24 {
		for _, value := range f {
			c := int(t) / value.value
			if c != 0 {
				return strconv.Itoa(c) + value.s + "前"
			}
		}
	}

	if hTime.Year() == nTime.Year() {
		return nTime.Format("m-d H:i")
	} else {
		return nTime.Format("Y-m-d H:i")
	}
}
