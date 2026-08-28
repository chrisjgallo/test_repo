package sim

import "testing"

// The screen and world sizes main.go launches with, so the wiring is checked at
// the proportions it actually runs at.
const (
	testScreenWidth  = 1700
	testScreenHeight = 1000
	testWorldWidth   = testScreenWidth * 3
	testWorldHeight  = testScreenHeight * 3
)

// TestNewWithoutAStarLeavesEmptySpace is the default every existing run relies
// on: no flag, nothing in the world but what you put there.
func TestNewWithoutAStarLeavesEmptySpace(t *testing.T) {
	s := New(testScreenWidth, testScreenHeight, testWorldWidth, testWorldHeight, false)

	if got := len(s.world.Objects); got != 0 {
		t.Errorf("want an empty world, got %d objects: %+v", got, s.world.Objects)
	}
}

// TestNewWithAStarCentersIt covers the one line joining the -star flag to the
// world. The star has to land in the middle of the world rather than the middle
// of the screen -- they are not the same point, the world being three times the
// screen in each direction -- and the camera starts there too, so a launch with
// the flag opens looking straight at it.
func TestNewWithAStarCentersIt(t *testing.T) {
	s := New(testScreenWidth, testScreenHeight, testWorldWidth, testWorldHeight, true)

	if len(s.world.Objects) != 1 {
		t.Fatalf("want just the star, got %d objects: %+v", len(s.world.Objects), s.world.Objects)
	}

	star := s.world.Objects[0]
	if !star.Fixed {
		t.Errorf("the star should be fixed in place, got %+v", star)
	}
	if star.X != s.world.Width()/2 || star.Y != s.world.Height()/2 {
		t.Errorf("want the star at the middle of the world (%v, %v), got (%v, %v)",
			s.world.Width()/2, s.world.Height()/2, star.X, star.Y)
	}
	if star.X != s.camera.centerX || star.Y != s.camera.centerY {
		t.Errorf("the view should open on the star: star at (%v, %v), camera at (%v, %v)",
			star.X, star.Y, s.camera.centerX, s.camera.centerY)
	}
}

// TestStatsHoldTheStarApart pins the readout's policy: the star is reported on
// its own line and left out of the three describing what the run has spawned,
// rather than being counted in some of them and not others.
func TestStatsHoldTheStarApart(t *testing.T) {
	s := New(testScreenWidth, testScreenHeight, testWorldWidth, testWorldHeight, true)
	s.world.Spawn(100, 100)
	s.world.Spawn(200, 200)

	stats := s.spawnedStats()

	if stats.count != 2 {
		t.Errorf("want 2 spawned objects counted, got %d", stats.count)
	}
	if stats.starMass <= 0 {
		t.Errorf("want the star's mass reported on its own, got %v", stats.starMass)
	}
	if stats.mass >= stats.starMass {
		t.Errorf("two spawned objects (%v) should not outweigh the star (%v) -- "+
			"the star looks to be counted in both", stats.mass, stats.starMass)
	}
	// A fresh spawn is the smallest thing there is, and the star is by far the
	// largest, so anything star-sized here means the star was counted.
	if stats.largest > 1 {
		t.Errorf("want the largest spawned radius, got %v", stats.largest)
	}
}

// TestStatsWithoutAStarReportNoStar keeps the Star line out of a run that has
// none, rather than printing a zero.
func TestStatsWithoutAStarReportNoStar(t *testing.T) {
	s := New(testScreenWidth, testScreenHeight, testWorldWidth, testWorldHeight, false)
	s.world.Spawn(100, 100)

	if got := s.spawnedStats().starMass; got != 0 {
		t.Errorf("want no star mass in a starless world, got %v", got)
	}
}
