package adclear

import (
	"testing"
)

func TestClear(t *testing.T) {
	err := clear("C:\\Users\\Steve\\Videos\\PlayOn\\The Ark\\Season 1", "c:\\artifacts3", true)
	if err != nil {
		t.Fatal(err)
	}
}
