package time

import (
	"github.com/golang-module/carbon"
	"strconv"
	"time"
)

// DiffForHumans 人性化时间
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

// DiffForHumansVariety 多样化的时间整理
func DiffForHumansVariety(Time1 time.Time, String1 string, Time2 time.Time, String2 string) string {
	carbonTime1 := carbon.Time2Carbon(Time1)
	carbonTime2 := carbon.Time2Carbon(Time2)
	if carbonTime2.Gt(carbonTime1) {
		return String2 + "于: " + DiffForHumans(Time2)
	} else {
		return String1 + "于: " + DiffForHumans(Time1)
	}
}
