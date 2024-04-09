package util

import "encoding/json"

type Obj map[string]any
type Objs []Obj

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
