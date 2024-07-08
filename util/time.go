package util

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

var TimestamptzFormat = time.RFC3339
var DateFormat = time.DateOnly
var DBTimestamptzFormat = "2006-01-02 15:04:05Z07"

type TZ string // +08:00

type AggLevel string

const (
	LevelNA    AggLevel = "n/a"
	LevelHour  AggLevel = "hour"
	LevelDay   AggLevel = "day"
	LevelWeek  AggLevel = "week"
	LevelMonth AggLevel = "month"
)

func (a AggLevel) ToAggly() (res string) {
	switch a {
	case LevelHour, LevelWeek, LevelMonth:
		return string(a) + "ly"
	case LevelDay:
		return "daily"
	default:
		return string(a)
	}
}

type TsPoint struct {
	T time.Time `json:"t" db:"t"`
	V float64   `json:"v" db:"v"`
}

type TsPoints []TsPoint

func (t TsPoints) Sort() {
	sort.Slice(t, func(i, j int) bool {
		return t[i].T.Before(t[j].T)
	})
}

func DatesBetweenIncl(a, b time.Time) (ret []time.Time) {
	a = ToDate(a)
	b = ToDate(b)

	for s := a; !s.After(b); s = s.AddDate(0, 0, 1) {
		ret = append(ret, s)
	}
	return
}

func HrBtw(t time.Time, min, max int) bool {
	return t.Hour() <= max && t.Hour() >= min
}

func ParseTime(s string) (ret time.Time, err error) {
	if ret, err = time.Parse(time.RFC3339Nano, s); err == nil {
		return
	}

	if ret, err = time.Parse(time.RFC3339, s); err == nil {
		return
	}

	if ret, err = time.Parse(time.DateTime, s); err == nil {
		return
	}

	if ret, err = time.Parse(time.DateOnly, s); err == nil {
		return
	}
	return
}

func RandInDayStr(dateStr string) time.Time {
	t, err := time.Parse(DateFormat, dateStr)
	Panic(err)

	return t.Add(time.Second * time.Duration(rand.Intn(3600*16)))
}

// TimeEqual compare if two time is the same without considering the nano second
func TimeEqual(a, b time.Time) bool {
	return a.UTC().Format(time.RFC3339) == b.UTC().Format(time.RFC3339)
}

func ToDate(t time.Time) time.Time {
	return Truncate(t, LevelDay)
}

func ToDateUTC(t time.Time) time.Time {
	y, m, d := t.Date()
	return DateUTC(y, int(m), d)
}

func IsStartOfMonth(a time.Time) bool {
	startOfMonth := StartOfMonth(a)
	return startOfMonth == a
}

func IsLastDayOfMonth(a time.Time) bool {
	// Get the date of the next day
	nextDay := a.AddDate(0, 0, 1)

	// If the month of the next day is different from the current month,
	// it means the current day is the last day of the month
	return nextDay.Month() != a.Month()
}

func StartOfPrevMonth(a time.Time) time.Time {
	/*
		Note that StartOfMonth(a.AddDate(0, -1, 0)) is not guaranteed to give
		us the start date of the previous month (because if a = 30th March'23
		for instance, then subtracting one month from it gives us 2nd March'23)
	*/
	return StartOfMonth(a).AddDate(0, -1, 0)
}

func StartOfWeek(a time.Time) time.Time {
	return Truncate(a, LevelWeek)
}

func StartOfMonth(a time.Time) time.Time {
	return Truncate(a, LevelMonth)
}

func StartOfDay(a time.Time) time.Time {
	return Truncate(a, LevelDay)
}

func EndOfDay(a time.Time) time.Time {
	d, y, m := a.Date()
	return time.Date(d, y, m, 23, 59, 59, 999999999, a.Location())
}

func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func MinTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// end of month
func EndOfMonthOrYesterday(a time.Time) time.Time {
	return MinTime(EndOfMonth(a), DateFromTodayUTC(-1))
}

// end of month
func EndOfWeekOrYesterday(a time.Time) time.Time {
	return MinTime(EndOfWeek(a), DateFromTodayUTC(-1))
}

func EndOfWeek(a time.Time) time.Time {
	return StartOfWeek(a).AddDate(0, 0, 7).Add(time.Nanosecond * -1)
}

func EndOfMonth(a time.Time) time.Time {
	return StartOfMonth(a).AddDate(0, 1, 0).Add(time.Nanosecond * -1)
}

func SameDate(a, b time.Time) bool {
	a = a.UTC()
	b = b.UTC()

	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()

	return y1 == y2 && m1 == m2 && d1 == d2
}

func RandomMomentInADay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, rand.Intn(23), rand.Intn(59), rand.Intn(59), rand.Intn(10000), t.Location())
}

func adjustedWeekday(t time.Time) int {
	w := t.Weekday()
	if w == 0 {
		return 7
	}

	return int(w)
}

func Truncate(t time.Time, timeLevel AggLevel) time.Time {
	y, m, d := t.Date()
	switch timeLevel {
	case LevelHour:
		return time.Date(y, m, d, t.Hour(), 0, 0, 0, t.Location())
	case LevelDay:
		return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
	case LevelWeek:
		diff := adjustedWeekday(t) - 1
		return time.Date(y, m, d, 0, 0, 0, 0, t.Location()).AddDate(0, 0, -int(diff))
	case LevelMonth:
		return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
	default:
		return t
	}
}

func PartitionTruncate(t time.Time, timeLevel AggLevel) time.Time {
	switch timeLevel {
	case LevelDay:
		y, m, _ := t.Date()
		return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
	default:
		return time.Time{}
	}
}

type TsRange struct {
	Start    time.Time
	End      time.Time
	Order    string // asc & desc or empty
	AggLevel AggLevel
}

func (t TsRange) TimesAny() (ret []any, err error) {
	ts, err := t.Times()
	if err != nil {
		return
	}
	for _, t := range ts {
		ret = append(ret, t)
	}
	return
}

func TimeIncr(t time.Time, aggLevel AggLevel) (time.Time, error) {
	switch aggLevel {
	case LevelHour:
		return t.Add(time.Hour), nil
	case LevelWeek:
		return t.AddDate(0, 0, 7), nil
	case LevelDay:
		return t.AddDate(0, 0, 1), nil
	case LevelMonth:
		return t.AddDate(0, 1, 0), nil
	default:
		return time.Time{}, errors.New("invalid agg level")
	}
}

func (t TsRange) TimesBeforeTodayDesc() (ret []time.Time, err error) {
	ts, err := t.Times()
	if err != nil {
		return
	}

	now := time.Now()
	for i := range ts {
		if ts[i].After(now) {
			continue
		}
		ret = append(ret, ts[i])
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].After(ret[j]) })
	return
}

func (t TsRange) Times() (ret []time.Time, err error) {
	start := Truncate(t.Start, t.AggLevel)
	end := Truncate(t.End, t.AggLevel)

	cur := start

	for !cur.After(end) {
		ret = append(ret, cur)

		cur, err = TimeIncr(cur, t.AggLevel)
		if err != nil {
			return
		}
	}
	switch t.Order {
	case "asc":
		sort.Slice(ret, func(i, j int) bool { return ret[i].Before(ret[j]) })
	case "desc":
		sort.Slice(ret, func(i, j int) bool { return ret[i].After(ret[j]) })
	}
	return
}

func DateFromToday(daysOffset int, tz TZ) time.Time {
	today := ToDate(NowByTZ(tz))
	return today.AddDate(0, 0, daysOffset)
}

func MonthFromToday(monthOffset int) time.Time {
	thisMonth := Truncate(time.Now().UTC(), LevelMonth)
	return thisMonth.AddDate(0, monthOffset, 0)
}

func MonthsFromToday(start, end int) (ret []time.Time) {
	if end < start {
		return
	}

	startDate := MonthFromToday(start)
	endDate := MonthFromToday(end)
	for cur := startDate; !cur.After(endDate); cur = cur.AddDate(0, 1, 0) {
		ret = append(ret, cur)
	}
	return
}

func DatesFromToday(start, end int, tz TZ) (ret []time.Time) {
	if end < start {
		return
	}

	startDate := DateFromToday(start, tz)
	endDate := DateFromToday(end, tz)

	ret = Dates(startDate, endDate)
	return
}

func Dates(startDate, endDate time.Time) (ret []time.Time) {
	for cur := startDate; !cur.After(endDate); cur = cur.AddDate(0, 0, 1) {
		ret = append(ret, cur)
	}
	return
}

func ParseTZ(t time.Time) (ret TZ) {
	ret, _ = ToTZ(t.Format(time.RFC3339))
	return ret
}

func ToTZ(ts string) (ret TZ, err error) {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return
	}

	ret = TZ(t.Format("Z07:00"))
	if ret == "Z" {
		ret = TZ("+00:00")
	}
	return
}

func TimeByTz(t time.Time, tz TZ) time.Time {
	t = t.UTC()
	if tz == "Z" {
		tz = "+00:00"
	}
	polarity := tz[0]
	split := strings.Split(string(tz[1:]), ":")
	hr, _ := strconv.Atoi(split[0])
	mi, _ := strconv.Atoi(split[1])

	switch polarity {
	case '+':
		t = t.Add(time.Hour*time.Duration(hr) + time.Minute*time.Duration(mi))
	case '-':
		t = t.Add(-time.Hour*time.Duration(hr) - time.Minute*time.Duration(mi))
	default:
		panic(fmt.Errorf("wrong polarity: %b", polarity))
	}

	format := "2006-01-02T15:04:05"
	t, err := time.Parse(time.RFC3339, fmt.Sprintf("%s%s", t.Format(format), tz))
	if err != nil {
		panic(fmt.Errorf("error parse date in NowByTZ, tz: %s, t: %v", tz, t))
	}
	return t
}

func NowByTZ(tz TZ) time.Time {
	switch tz {
	case "", "Z":
		tz = "+00:00"
	}
	return TimeByTz(time.Now().UTC(), tz)
}

func YesterdayByTZ(tz TZ) time.Time {
	d := NowByTZ(tz).AddDate(0, 0, -1)
	return ToDate(d) // UTC here has no meaning, because golang doesn't have date type, we use UTC as default placeholder
}

func YesterdayOf(t time.Time) time.Time {
	return ToDate(t).AddDate(0, 0, -1)
}

func DateFromTodayUTC(offset int) time.Time {
	return Truncate(time.Now().UTC(), LevelDay).AddDate(0, 0, offset)
}

func LastMonth() time.Time {
	y, m, _ := time.Now().UTC().AddDate(0, -1, 0).Date()
	return DateUTC(y, int(m), 1)
}

func MustTime(str string) time.Time {
	t, err := time.Parse(time.RFC3339, str)
	Panic(err)
	return t
}

func DateUTCFromTime(t time.Time) time.Time {
	y, m, d := t.Date()
	return DateUTC(y, int(m), d)
}

func DateUTC(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC) // the utc here is meaningless
}

// random moment in the same date
func RM(t time.Time) time.Time {
	return ToDate(t).Add(time.Second * time.Duration(rand.Intn(3600*23)))
}

func MaxDayInMonth(month int) (day int) {
	nonLeapYear := 2021
	// a day value of 0 gives the last day of the previous month
	maxDay := time.Date(nonLeapYear, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	day = rand.Intn(maxDay) + 1
	return
}

func RandomizeTime(t time.Time, parts string) time.Time {
	parts = strings.ToLower(parts)
	components := strings.Split(parts, "")

	year := t.Year()
	month := t.Month()
	day := t.Day()
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()
	nanosecond := t.Nanosecond()

	for _, component := range components {
		switch component {
		case "y":
			year = year - 3 + rand.Intn(4)
		case "m":
			month = time.Month(rand.Intn(12) + 1)
		case "d":
			day = MaxDayInMonth(int(month))
		case "h":
			hour = rand.Intn(24)
		case "i":
			minute = rand.Intn(60)
		case "s":
			second = rand.Intn(60)
		case "n":
			nanosecond = rand.Intn(1000000000)
		}
	}

	return time.Date(year, month, day, hour, minute, second, nanosecond, t.Location())
}

func Timestamp(y, m, d, h, min, s int) time.Time {
	return time.Date(y, time.Month(m), d, h, min, s, 0, time.UTC)
}

func LastDayOfTheMonth(date time.Time) time.Time {
	month := Truncate(date, LevelMonth)
	return month.AddDate(0, 1, 0).AddDate(0, 0, -1)
}

func LastDayOfTheWeek(date time.Time) time.Time {
	date = Truncate(date, LevelWeek)
	return date.AddDate(0, 0, 6)
}

var log = NewLogger("util.latency")

func Latency(marker time.Time, kvs map[string]any) {
	pc, _, _, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	l := log.Info().Dur("dur", time.Since(marker)).
		Str("caller", details.Name())

	for k, v := range kvs {
		l.Interface(k, v)
	}

	l.Msg("##")
}
