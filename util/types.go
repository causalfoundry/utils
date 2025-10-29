package util

import (
	"encoding/json"
	"errors"
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

func (o *Objs) Scan(value any) error {
	if value == nil {
		*o = Objs{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot marshal to objs")
	}
	return json.Unmarshal(bytes, o)
}

func (o *Obj) Scan(value any) error {
	if value == nil {
		*o = Obj{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot marshal to objs")
	}
	return json.Unmarshal(bytes, o)
}
