// +build gofuzz

package pb_fuzz_workshop

import (
	"encoding/json"
	"reflect"
)

func Fuzz(b []byte) int {
	var s []int
	if err := json.Unmarshal(b, &s); err != nil {
		return 0
	}

	if reflect.DeepEqual(Reverse(s), s) {
		return 0
	}

	if !reflect.DeepEqual(Reverse(Reverse(s)), s) {
		panic("not equal")
	}

	return 1
}
