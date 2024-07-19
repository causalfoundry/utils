package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRange(t *testing.T) {
	yd := YesterdayByTZ("+08:00")
	rg := TsRange{
		Start:    yd.AddDate(0, 0, -2),
		End:      yd,
		Order:    "desc",
		AggLevel: "day",
	}
	ts, err := rg.TimesBeforeToday()
	assert.Nil(t, err)
	ForEach(ts, func(t time.Time) {
		fmt.Println(t)
	})

	rg.Order = "asc"
	ts, err = rg.Times()
	assert.Nil(t, err)
	assert.True(t, ts[0].Before(ts[1]))

	rg.Order = "desc"
	ts, err = rg.Times()
	assert.Nil(t, err)
	assert.False(t, ts[0].Before(ts[1]))
}

func TestTruncate(t *testing.T) {
	ti, err := time.Parse(time.RFC3339, "2020-01-01T04:20:20+08:00")
	assert.Nil(t, err)
	newTi := Truncate(ti, LevelHour)
	assert.Equal(t, newTi.Format(time.RFC3339), "2020-01-01T04:00:00+08:00")

	for i := 0; i < 20; i++ {
		ti = Truncate(ti.AddDate(0, 0, i), LevelWeek)
		assert.Equal(t, ti.Weekday(), time.Monday)
	}
}

func TestRandInDayStr(t *testing.T) {
	rt := RandInDayStr("2020-01-03")
	fmt.Println(rt)
}

func TestTimesOfTsRange(t *testing.T) {
	ts := TsRange{
		Start:    DateUTC(2020, 1, 4),
		End:      DateUTC(2020, 2, 3),
		AggLevel: LevelMonth,
		Order:    "asc",
	}

	tis, err := ts.Times()
	assert.Nil(t, err)
	assert.Equal(t, tis, []time.Time{DateUTC(2020, 1, 1), DateUTC(2020, 2, 1)})

	ts.AggLevel = LevelDay
	tis, err = ts.Times()
	assert.Nil(t, err)
	assert.Len(t, tis, 31-4+1+3)

	ts = TsRange{
		Start:    time.Now().AddDate(0, 0, -4),
		End:      time.Now().AddDate(0, 0, 3),
		Order:    "desc",
		AggLevel: LevelDay,
	}
	tss, err := ts.TimesBeforeToday()
	assert.Nil(t, err)
	for i := 1; i < len(tss); i++ {
		assert.True(t, tss[i].Before(tss[i-1]))
	}

	t.Run("weekly", func(t *testing.T) {
		ts := TsRange{
			Start:    DateUTC(2024, 1, 15),
			End:      DateUTC(2024, 1, 22), // the following monday
			AggLevel: LevelWeek,
			Order:    "asc",
		}

		times, err := ts.Times()
		assert.Nil(t, err)
		assert.Equal(t, times, []time.Time{DateUTC(2024, 1, 15), DateUTC(2024, 1, 22)})

		ts = TsRange{
			Start:    DateUTC(2024, 1, 15),
			End:      DateUTC(2024, 1, 17), // the following monday
			AggLevel: LevelWeek,
			Order:    "asc",
		}

		times, err = ts.Times()
		assert.Nil(t, err)
		assert.Equal(t, times, []time.Time{DateUTC(2024, 1, 15)})
	})
}

func TestMisc(t *testing.T) {
	ts := DateUTC(2024, 1, 8) // monday
	ret := EndOfWeekOrYesterday(ts)
	assert.True(t, ret.Equal(DateUTC(2024, 1, 15).Add(time.Nanosecond*-1)))
}

func TestTimeLocation(t *testing.T) {
	// Parse the RFC 3339 time string.
	timeStr := "2020-01-01T00:00:00+08:00"
	tt, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}

	// Get the Location of the parsed time.
	loc := tt.Location()

	// Print the Location's string representation.
	fmt.Println("--", loc.String())
	name, offset := tt.Zone()
	fmt.Println(name, offset)
}

func TestToTZ(t *testing.T) {
	ts0 := "2020-01-03T04:00:03Z"
	ts1 := "2020-01-03T04:00:03+00:00"
	ts2 := "2020-01-03T04:00:03+05:00"
	tz, err := ToTZ(ts0)
	assert.Nil(t, err)
	assert.Equal(t, tz, TZ("+00:00"))

	tz, err = ToTZ(ts1)
	assert.Nil(t, err)
	assert.Equal(t, tz, TZ("+00:00"))

	tz, err = ToTZ(ts2)
	assert.Nil(t, err)
	assert.Equal(t, tz, TZ("+05:00"))
}

func TestYesterdayTZ(t *testing.T) {
	tt := YesterdayByTZ("+08:00")
	fmt.Println(tt)
	fmt.Println(tt.Format(time.RFC3339))
	fmt.Println(tt.UTC())

	tt = YesterdayByTZ("+00:00")
	fmt.Println(tt)
	fmt.Println(tt.Format(time.RFC3339))
	fmt.Println(tt.UTC())
}

func TestDatesFromToDay(t *testing.T) {
	dates := DatesFromToday(-3, -1, "+00:00")
	assert.Len(t, dates, 3)
}

func TestMonthesFromToday(t *testing.T) {
	dates := MonthsFromToday(-3, -1)
	assert.Len(t, dates, 3)
}

func TestTsPoints(t *testing.T) {
	ts := TsPoints{
		{
			T: DateUTC(2020, 3, 1),
		},
		{
			T: DateUTC(2020, 2, 1),
		},
		{
			T: DateUTC(2020, 4, 1),
		},
	}
	expected := TsPoints{
		{
			T: DateUTC(2020, 2, 1),
		},
		{
			T: DateUTC(2020, 3, 1),
		},
		{
			T: DateUTC(2020, 4, 1),
		},
	}
	ts.Sort()

	assert.Equal(t, ts, expected)
}
