package term

import (
	"testing"
)

func TestTermSettings(t *testing.T) {
	tio, err := NewTermSettings(0)
	if err != nil {
		t.Errorf("NewTermSettings: %s", err)
	}
	if err := tio.apply(); err != nil {
		t.Errorf("apply: %s", err)
	}
	if err := tio.Reset(); err != nil {
		t.Errorf("Reset: %s", err)
	}
	t.Errorf("%v",tio)
}
