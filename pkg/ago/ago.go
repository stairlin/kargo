package ago

import (
	"fmt"
	"math"
	"time"
)

// Source: https://github.com/ararog/timeago/blob/master/timeago.go

type unit int

const (
	second unit = iota
	minute
	hour
	day
	week
	month
	year
)

// Ago returns the distance in time between t and now in words
// e.g. 3 weeks ago
func Ago(t time.Time) string {
	dist := time.Now().Sub(t)
	if dist.Hours() < 24 {
		if dist.Hours() >= 1 {
			return sprint(hour, int(round(dist.Hours())))
		} else if dist.Minutes() >= 1 {
			return sprint(minute, int(round(dist.Minutes())))
		} else {
			return sprint(second, int(round(dist.Seconds())))
		}
	} else {
		if dist.Hours() >= 8760 {
			years := dist.Hours() / 8760
			return sprint(year, int(years))
		} else if dist.Hours() >= 730 {
			months := dist.Hours() / 730
			return sprint(month, int(months))
		} else if dist.Hours() >= 168 {
			weeks := dist.Hours() / 168
			return sprint(week, int(weeks))
		}
		days := dist.Hours() / 24
		return sprint(day, int(days))
	}
}

func sprint(valueType unit, value int) string {
	switch valueType {
	case year:
		if value >= 2 {
			return fmt.Sprintf("%d years ago", value)
		}
		return "Last year"
	case month:
		if value >= 2 {
			return fmt.Sprintf("%d months ago", value)
		}
		return "Last month"
	case week:
		if value >= 2 {
			return fmt.Sprintf("%d weeks ago", value)
		}
		return "Last week"
	case day:
		if value >= 2 {
			return fmt.Sprintf("%d days ago", value)
		}
		return "Yesterday"
	case hour:
		if value >= 2 {
			return fmt.Sprintf("%d hours ago", value)
		}
		return "An hour ago"
	case minute:
		if value >= 2 {
			return fmt.Sprintf("%d minutes ago", value)
		}
		return "A minute ago"
	case second:
		if value >= 2 {
			return fmt.Sprintf("%d seconds ago", value)
		}
		return "Just now"
	}
	return ""
}

func round(f float64) float64 {
	return math.Floor(f + .5)
}
