package term

import (
	"testing"
)

func TestTermSettings(t *testing.T) {
	tio, err := NewTermSettings(0)
	if err != nil {
		t.Errorf("NewTermSettings: %s", err)
	}
	if err := tio.Apply(); err != nil {
		t.Errorf("Apply: %s", err)
	}
	if err := tio.Reset(); err != nil {
		t.Errorf("Reset: %s", err)
	}
	t.Log(tio)
}
