package util

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Obj map[string]any
type Objs []Obj
type NullFloat64 struct {
	Null bool
	V    float64
}

func (o *NullFloat64) Scan(v any) error {
	*o = NullFloat64{}
	if v == nil {
		o.Null = true
		return nil
	}

	f, err := strconv.ParseFloat(fmt.Sprint(v), 64)
	if err != nil {
		return err
	}
	(*o).V = f
	return nil
}

func (o NullFloat64) MarshalJSON() ([]byte, error) {
	if o.Null {
		return []byte("null"), nil
	}
	return json.Marshal(o.V)
}

func (o *Objs) Scan(value interface{}) error {
	if value == nil {
		*o = Objs{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return ErrNotBytes
	}
	return json.Unmarshal(bytes, o)
}

func (o *Obj) Scan(value interface{}) error {
	if value == nil {
		*o = Obj{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return ErrNotBytes
	}
	return json.Unmarshal(bytes, o)
}
