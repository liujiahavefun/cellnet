package util

import (
	"testing"
)

func TestStrHash(t *testing.T) {
	if v := StringHash("gamedef.EnterGameREQ"); v != 0x47c9ce66 {
		t.Errorf("expect 0x47c9ce66, got 0x%x", v)
	}

	if v := StringHash("gamedef.EnterGameACK"); v != 0x2c933204 {
		t.Errorf("expect 0x2c933204, got 0x%x", v)
	}

	if v := StringHashNoCase("gamedef.EnterGameREQ"); v != 0x47c9ce66 {
		t.Errorf("expect 0x47c9ce66, got 0x%x", v)
	}

	if v := StringHashNoCase("gamedef.EnterGameACK"); v != 0x2c933204 {
		t.Errorf("expect 0x2c933204, got 0x%x", v)
	}
}
