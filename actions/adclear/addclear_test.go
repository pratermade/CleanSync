package adclear

import (
	"testing"
)

func TestClear(t *testing.T) {
	err := clear("c:\\artifacts\\original.mp4", "x:\\artifacts\\new.mp4", true)
	if err != nil {
		t.Fatal(err)
	}
}
