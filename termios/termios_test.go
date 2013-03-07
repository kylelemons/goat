// Run this test with:
//   go test -c && ./term.test

package termios

import (
	"testing"
)

func TestTermSettings(t *testing.T) {
	tio, err := NewTermSettings(0)
	if err != nil {
		t.Fatalf("NewTermSettings: %s", err)
	}
	if err := tio.Apply(); err != nil {
		t.Errorf("Apply: %s", err)
	}
	if err := tio.Reset(); err != nil {
		t.Errorf("Reset: %s", err)
	}
	t.Log(tio)
}

func TestTermSize(t *testing.T) {
	tio, err := NewTermSettings(0)
	if err != nil {
		t.Fatalf("NewTermSettings: %s", err)
	}
	w, h, err := tio.GetSize()
	if err != nil {
		t.Fatalf("GetSize: %s", err)
	}
	t.Logf("Size: %d cols, %d rows", w, h)
}
