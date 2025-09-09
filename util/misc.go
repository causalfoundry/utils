package util

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func ParseJSONToStruct(src, dest any) error {
	var data []byte
	if src == nil {
		return nil
	}

	if b, ok := src.([]byte); ok {
		data = b
	} else {
		return ErrNotBytes
	}

	return json.Unmarshal(data, dest)
}

func InBracket(s, left, right string) bool {
	return strings.HasPrefix(s, left) && strings.HasSuffix(s, right)
}

func OrDefault(value string, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

func ToFloat64(x any) (val float64, err error) {
	switch v := x.(type) {
	case int:
		val = float64(v)
	case float64:
		val = v
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		err = fmt.Errorf("error in converting type %T to float64, %v", x, x)
	}
	return val, err
}

func Print(obj ...any) {
	for _, o := range obj {
		if reflect.ValueOf(o).Kind() == reflect.Struct {
			b, _ := json.MarshalIndent(o, "", "  ")
			fmt.Printf("%s\n", string(b))
		} else {
			fmt.Printf("%+v\n", o)
		}
	}
}

// assume first letter is upper
func CamelToSnake2(s string) string {
	var idx, accIdx int
	var tmp []string
	var group string
	for {
		if accIdx >= len(s) {
			return strings.Join(tmp, "_")
		}

		group, idx = findNextGroup(s[accIdx:])
		accIdx += idx
		tmp = append(tmp, strings.ToLower(group))
	}
}

func findNextGroup(s string) (string, int) {
	var left = 0
	var right = -1

	if len(s) == 0 {
		return "", 0
	}

	if unicode.IsUpper(rune(s[0])) {
		// first letter is upper
		for i := left + 1; i < len(s); i++ {
			if unicode.IsUpper(rune(s[i])) && unicode.IsLower(rune(s[i-1])) {
				right = i
				break
			}
		}

	} else {
		// first letter is lower
		for i, r := range s {
			if unicode.IsUpper(r) {
				right = i
				break
			}
		}
	}

	if right < 0 {
		right = len(s)
	}
	return s[left:right], right
}

func CamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func ContainInt(sl []int, elem int) bool {
	for _, s := range sl {
		if s == elem {
			return true
		}
	}
	return false
}

func NewRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandomAlphanumeric(l int, lowerCase bool) string {
	b := strings.Builder{}
	al := len(alphanumeric)
	r := NewRand()
	for i := 0; i < l; i++ {
		_ = b.WriteByte(alphanumeric[r.Intn(al)])
	}
	if lowerCase {
		return strings.ToLower(b.String())
	}
	return b.String()
}

func ArrGet[T any](arr []T, idx []int) (ret []T) {
	for _, i := range idx {
		ret = append(ret, arr[i])
	}
	return
}

func ToObj(s any) (ret Obj) {
	b, _ := json.Marshal(s)
	err := json.Unmarshal(b, &ret)
	Panic(err)
	return
}

func JoinByComma[T any](t []T) string {
	ret := make([]string, len(t))
	for i := range t {
		ret[i] = fmt.Sprint(t[i])
	}
	return strings.Join(ret, ", ")
}

func ReadCSVFile(filename string) (map[string][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	headers := records[0]
	df := make(map[string][]string)
	for _, row := range records[1:] {
		for index, col := range row {
			headerName := headers[index]
			df[headerName] = append(df[headerName], col)
		}
	}
	return df, nil
}

func ToStr[T any](a T) string {
	switch v := any(a).(type) {
	case time.Time:
		return v.Format(time.RFC3339)
	case string:
		return fmt.Sprintf("'%s'", v)
	default:
		return fmt.Sprint(a)
	}
}

func DBJsonScan[T any](in any) (ret T, err error) {
	if in == nil {
		return
	}
	bytes, ok := in.([]byte)
	if !ok {
		err = ErrNotBytes
		return
	}
	err = json.Unmarshal(bytes, &ret)
	return
}
