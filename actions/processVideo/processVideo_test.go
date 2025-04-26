package processVideo

import (
	"testing"
)

func TestClear(t *testing.T) {
	err := clear("C:\\Users\\Steve\\Videos\\PlayOn\\The Big Bang Theory\\Season 11", "c:\\artifacts3", true)
	if err != nil {
		t.Fatal(err)
	}
}
