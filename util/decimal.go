package util

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

type Decimal decimal.Decimal

func (d Decimal) Unwrap() decimal.Decimal {
	return decimal.Decimal(d)
}

func (d Decimal) String() string {
	return decimal.Decimal(d).String()
}

func NewDecimalF(f float64) Decimal {
	return Decimal(decimal.NewFromFloat(f))
}

func NewDecimal(s string) Decimal {
	v, err := decimal.NewFromString(s)
	Panic(err)
	return Decimal(v)
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	return []byte(`"` + decimal.Decimal(d).String() + `"`), nil
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		return nil
	}

	if str[0] != '"' || str[len(str)-1] != '"' {
		return errors.New("decimal value needs to be quote as string")
	}

	v, err := decimal.NewFromString(str[1 : len(str)-1])
	if err != nil {
		return err
	}

	*d = Decimal(v)
	return nil
}

func (o *Decimal) Scan(value any) (err error) {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot convert to string: %v, %T", value, value)
	}
	d, err := decimal.NewFromString(str)
	*o = Decimal(d)
	return
}

func (o Decimal) Value() (driver.Value, error) {
	return decimal.Decimal(o).String(), nil
}
