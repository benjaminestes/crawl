package crawler

import (
	"os"
	"testing"
)

func TestIllFormed(t *testing.T) {
	f, err := os.Open("testdata/ill_formed.json")
	if err != nil {
		t.Fatalf("couldn't open test data")
	}
	defer f.Close()

	_, err = FromJSON(f)
	if err == nil {
		t.Fatalf("invalid config file should trigger error\n")
	}
}

func TestBadWait(t *testing.T) {
	f, err := os.Open("testdata/bad_wait.json")
	if err != nil {
		t.Fatalf("couldn't open test data")
	}
	defer f.Close()

	c, err := FromJSON(f)
	if err != nil {
		t.Fatalf("invalid config file shouldn't trigger immediate error\n")
	}

	err = c.Start()
	if err == nil {
		t.Fatalf("invalid config should trigger error on start")
	}
}
