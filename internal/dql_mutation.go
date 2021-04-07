package internal

import (
	"fmt"
	"strings"
	"time"
)

type DqlMutation struct {
	data []string
}

func NewDqlMutation() *DqlMutation {
	return new(DqlMutation)
}

func (d *DqlMutation) addIf(cond bool, k string, v interface{}) *DqlMutation {
	if cond {
		return d.add(k, v)
	} else {
		return d
	}
}
func (d *DqlMutation) add(k string, v interface{}) *DqlMutation {
	var val string
	switch t := v.(type) {
	case int, int8, int16, int32, int64:
		val = fmt.Sprintf("%d", t)
		break
	case float32, float64:
		val = fmt.Sprintf("%f", t)
		break
	case time.Time:
		val = `"` + v.(time.Time).Format(time.RFC3339) + `"`
		break
	case nil:
		val = "null"
		break
	case string:
		if len(v.(string)) > 0 && (v.(string)[0:1] == "{" || v.(string)[0:1] == "[") {
			val = v.(string)
		} else {
			val = `"` + v.(string) + `"`
		}
		break
	case bool:
		if v.(bool) {
			val = "true"
		} else {
			val = "false"

		}
		break
	default:
		fmt.Printf("I don't know about type %T!\n", v)
		val = string(v.(string))
		break
	}
	d.data = append(d.data, `"`+k+`":`+val)
	return d
}
func (d *DqlMutation) serialize() string {
	return `[{` + strings.Join(d.data, ",") + `}]`
}
