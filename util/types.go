package util

import "encoding/json"

type Obj map[string]any

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
