package katarproW

import "testing"

func TestPerformanceOptionsAreDeterministicAndFailClosed(t *testing.T) {
	options := katarProWOptions(map[int]string{5: "2000 Hz / 0.5 msec", 1: "125 Hz / 8 msec", 4: "1000 Hz / 1 msec"})
	if len(options) != 3 || options[0].Value != 1 || options[2].Value != 5 {
		t.Fatal("polling options were not sorted")
	}
	if katarProWOptions(map[int]string{1: ""}) != nil {
		t.Fatal("empty option label must fail closed")
	}
}
