package test

import (
	"testing"

	"github.com/neptulon/randstr"
)

var (
	msg1 = "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	msg2 = "In sit amet lectus felis, at pellentesque turpis."
	msg3 = "Nunc urna enim, cursus varius aliquet ac, imperdiet eget tellus."
	msg4 = randstr.Get(45 * 1000)       // 0.45 MB
	msg5 = randstr.Get(5 * 1000 * 1000) // 5.0 MB
)

func TestMessages(t *testing.T) {
	// todo: verify all message echoes from small to big
}

func TestBidirectional(t *testing.T) {
	// todo: test simultaneous read/writes
}
