package util

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDecimal(t *testing.T) {
	t.Run("", func(t *testing.T) {
		d := Decimal{}

		b, err := json.Marshal(d)
		assert.Equal(t, string(b), `"0"`)
		assert.Nil(t, err)
	})

	type mock struct {
		Name  string  `json:"name" db:"name"`
		Price Decimal `json:"price" db:"price"`
	}

	db := NewTestPostgresDB("")

	t.Run("json", func(t *testing.T) {
		actual := mock{
			Name:  "name",
			Price: NewDecimal("123.323"),
		}
		b, err := json.Marshal(actual)

		strJSON := `{"name":"name","price":"123.323"}`
		assert.Nil(t, err)
		assert.Equal(t, string(b), strJSON)

		var expect mock
		assert.Nil(t, json.Unmarshal([]byte(strJSON), &expect))
		assert.Equal(t, actual, expect)
	})

	t.Run("db", func(t *testing.T) {
		_, err := db.Exec("CREATE TABLE haha(name varchar, price decimal(10,2))")
		assert.Nil(t, err)

		m := mock{Name: "a", Price: NewDecimal("20.03")}

		_, err = db.Exec("INSERT INTO haha (name, price) VALUES ($1, $2)", m.Name, m.Price)
		assert.Nil(t, err)

		var ret []mock
		err = db.Select(&ret, "SELECT * FROM haha")
		assert.Nil(t, err)
		assert.Len(t, ret, 1)
		assert.Equal(t, ret[0], m)
	})
}

func TestDecimalSerialization(t *testing.T) {
	a := `212`
	var d Decimal
	e := json.Unmarshal([]byte(a), &d)
	assert.NotNil(t, e)

	a = `"212"`
	e = json.Unmarshal([]byte(a), &d)
	assert.Nil(t, e)
	assert.Equal(t, NewDecimalF(212).String(), d.String())
}

func TestDecimalScan(t *testing.T) {
	var d Decimal

	assert.Nil(t, d.Scan("12.34"))
	assert.Equal(t, "12.34", d.String())

	assert.Nil(t, d.Scan([]byte("56.78")))
	assert.Equal(t, "56.78", d.String())

	base := decimal.RequireFromString("90.12")
	assert.Nil(t, d.Scan(base))
	assert.Equal(t, "90.12", d.String())

	assert.Nil(t, d.Scan(Decimal(base)))
	assert.Equal(t, "90.12", d.String())

	assert.NotNil(t, d.Scan(nil))
}
