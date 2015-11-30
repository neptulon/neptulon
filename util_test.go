package neptulon

import "testing"

func TestRandString(t *testing.T) {
	l := 12304
	str := randString(l)

	if len(str) != l {
		t.Fatalf("Expected a random string of length %v but got %v", l, len(str))
	}
	if str[1] == str[2] && str[3] == str[4] && str[5] == str[6] && str[7] == str[8] {
		t.Fatal("Expected a random string, got repeated characters")
	}
}
