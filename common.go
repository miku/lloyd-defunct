package lloyd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const AppVersion = "0.2.4"

// StringValue returns the value for a given key in dot notation
func StringValue(key string, doc map[string]interface{}) (string, error) {
	keys := strings.Split(key, ".")
	if len(keys) == 0 {
		return "", fmt.Errorf("keys exhausted")
	}
	head := keys[0]
	val, ok := doc[head]
	if !ok {
		return "", fmt.Errorf("key %s not found", head)
	}
	switch t := val.(type) {
	case string:
		return val.(string), nil
	case map[string]interface{}:
		if len(keys) < 2 {
			return "", fmt.Errorf("no value found")
		}
		return StringValue(keys[1], val.(map[string]interface{}))
	case float64:
		return strconv.FormatFloat(val.(float64), 'f', 6, 64), nil
	case int:
		return strconv.Itoa(val.(int)), nil
	case json.Number:
		return fmt.Sprintf("%s", val), nil
	case []interface{}:
		return "", fmt.Errorf("not supported yet: %+v\n", t)
	default:
		return "", fmt.Errorf("unknown type: %+v, %+v\n", t, reflect.TypeOf(t))
	}
	return "", fmt.Errorf("no value found")
}
