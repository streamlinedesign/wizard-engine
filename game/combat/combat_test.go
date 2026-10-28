package combat

import "testing"

func TestProjectileHitsDummyOnce(t *testing.T) {
	var w World
	d := NewDummy(100, 100)
	w.Dummies = []*Dummy{d}
	w.Fire(60, 90, 1, 0, 100, 3, 10, 200, [3]uint8{255, 0, 0})
	open := func(x, y float64) bool { return false }
	for i := 0; i < 60; i++ {
		w.Update(1.0/60, open)
	}
	if d.HP != 30 {
		t.Fatalf("dummy should have taken 10 once, HP %v", d.HP)
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("projectile should be consumed on hit")
	}
	if len(w.Texts) == 0 {
		t.Fatal("expected a damage number")
	}
}

func TestProjectileStopsAtWallsAndRange(t *testing.T) {
	var w World
	w.Fire(0, 0, 1, 0, 100, 3, 1, 50, [3]uint8{})
	wall := func(x, y float64) bool { return x > 20 }
	for i := 0; i < 30; i++ {
		w.Update(1.0/60, wall)
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("projectile should die at the wall")
	}
	w.Fire(0, 0, 1, 0, 100, 3, 1, 50, [3]uint8{})
	for i := 0; i < 60; i++ {
		w.Update(1.0/60, func(x, y float64) bool { return false })
	}
	if len(w.Projectiles) != 0 {
		t.Fatal("projectile should die at range")
	}
}

func TestDummyFallsAndRespawns(t *testing.T) {
	d := NewDummy(0, 0)
	if !d.Hit(100) || d.Alive() {
		t.Fatal("should fall")
	}
	if d.Hit(5) {
		t.Fatal("a fallen dummy takes no hits")
	}
	for i := 0; i < 200; i++ {
		d.Update(1.0 / 60)
	}
	if !d.Alive() || d.HP != d.MaxHP {
		t.Fatal("should have respawned at full HP")
	}
}
