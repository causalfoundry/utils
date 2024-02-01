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
		AggLevel: "day",
	}
	ts, err := rg.TimesBeforeTodayDesc()
	assert.Nil(t, err)
	ForEach(ts, func(t time.Time) {
		fmt.Println(t)
	})
}

func TestTruncate(t *testing.T) {
	ti, err := time.Parse(time.RFC3339, "2020-01-01T04:20:20+08:00")
	assert.Nil(t, err)
	newTi := Truncate(ti, AggLevel(LevelHour))
	assert.Equal(t, newTi.Format(time.RFC3339), "2020-01-01T04:00:00+08:00")
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
		AggLevel: LevelDay,
	}
	tss, err := ts.TimesBeforeTodayDesc()
	assert.Nil(t, err)
	fmt.Println(tss)

	t.Run("weekly", func(t *testing.T) {
		ts := TsRange{
			Start:    DateUTC(2024, 1, 15),
			End:      DateUTC(2024, 1, 22), // the following monday
			AggLevel: LevelWeek,
		}

		times, err := ts.Times()
		assert.Nil(t, err)
		assert.Equal(t, times, []time.Time{DateUTC(2024, 1, 15), DateUTC(2024, 1, 22)})

		ts = TsRange{
			Start:    DateUTC(2024, 1, 15),
			End:      DateUTC(2024, 1, 17), // the following monday
			AggLevel: LevelWeek,
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
