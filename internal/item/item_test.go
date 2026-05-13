package item

import "testing"

func TestRegistry_BasicLookup(t *testing.T) {
	r := NewRegistry()
	bow, err := r.Get("Bow", 0)
	if err != nil {
		t.Fatalf("Get Bow: %v", err)
	}
	if bow.Name != "Bow" || bow.Type != TypeBow {
		t.Errorf("unexpected: %+v", bow)
	}
	if got := bow.FullName(); got != "Bow:0" {
		t.Errorf("FullName=%q, want Bow:0", got)
	}
}

func TestRegistry_AliasResolves(t *testing.T) {
	r := NewRegistry()
	us, err := r.Get("UncleSword", 0)
	if err != nil {
		t.Fatalf("Get UncleSword: %v", err)
	}
	if us.Type != TypeAlias {
		t.Errorf("expected TypeAlias, got %v", us.Type)
	}
	if us.Target == nil || us.Target.Name != "ProgressiveSword" {
		t.Errorf("alias target = %+v", us.Target)
	}
	// IsType should also resolve through alias to TypeSword.
	if !us.IsType(TypeSword) {
		t.Error("UncleSword should report TypeSword via alias")
	}
}

func TestRegistry_Medallion_NamedBytes(t *testing.T) {
	r := NewRegistry()
	bombos, err := r.Get("Bombos", 0)
	if err != nil {
		t.Fatalf("Get Bombos: %v", err)
	}
	nb := bombos.GetNamedBytes()
	want := map[string]int{"t0": 0x31, "t1": 0x90, "t2": 0x00, "m0": 0x31, "m1": 0x80, "m2": 0x00}
	for k, v := range want {
		if nb[k] != v {
			t.Errorf("Bombos[%s] = 0x%X, want 0x%X", k, nb[k], v)
		}
	}
	if bombos.Bytes[0] != 0x0F || bombos.Bytes[1] != 0x00 {
		t.Errorf("Bombos.Bytes = %v", bombos.Bytes)
	}
}

func TestRegistry_HeartPower(t *testing.T) {
	r := NewRegistry()
	poh, _ := r.Get("PieceOfHeart", 0)
	hc, _ := r.Get("HeartContainer", 0)
	if poh.Power != 0.25 {
		t.Errorf("PoH power = %v, want 0.25", poh.Power)
	}
	if hc.Power != 1 {
		t.Errorf("HC power = %v, want 1", hc.Power)
	}
}

func TestRegistry_PerWorldCache(t *testing.T) {
	r := NewRegistry()
	a, _ := r.Get("Bow", 0)
	b, _ := r.Get("Bow", 1)
	if a == b {
		t.Error("same item across worlds; should be distinct instances")
	}
	if a.WorldID != 0 || b.WorldID != 1 {
		t.Errorf("world IDs wrong: %d %d", a.WorldID, b.WorldID)
	}
}
