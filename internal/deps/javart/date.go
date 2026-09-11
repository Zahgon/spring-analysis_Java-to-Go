package javart

import "time"

// dateLayout is java.util.Date.toString's format: "EEE MMM dd HH:mm:ss zzz yyyy".
const dateLayout = "Mon Jan 02 15:04:05 MST 2006"

// DateToString renders a time the way java.util.Date.toString does, e.g.
// "Thu Jan 02 03:04:05 UTC 2020". A nil time renders as "null", the way a null
// Date field does inside SimpleModel.toString.
func DateToString(t *time.Time) string {
	if t == nil {
		return Null
	}
	return t.Format(dateLayout)
}
