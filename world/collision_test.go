package world

import (
	"math"
	"testing"
)

// The pair used throughout: two default objects eight apart on the x axis, well
// inside the collision band, with the normal pointing cleanly along +x.
const testDistance = 8

func TestASolidHitBouncesBothObjectsApart(t *testing.T) {
	a := SpaceObject{X: 496, Y: 500, Radius: defaultRadius, VelocityX: 20, Mass: defaultMass}
	b := SpaceObject{X: 504, Y: 500, Radius: defaultRadius, VelocityX: -20, Mass: defaultMass}

	collide(&a, &b, testDistance)

	if a.Mass == 0 || b.Mass == 0 {
		t.Fatal("a hit this hard should bounce, not merge")
	}

	// Equal masses closing at 40, so each leaves at restitution * 20.
	assertClose(t, "a velocity", a.VelocityX, -restitution*20)
	assertClose(t, "b velocity", b.VelocityX, restitution*20)
}

func TestBounceConservesMomentumAndShedsSpeed(t *testing.T) {
	// Something light hitting something three times heavier, so a mistake in
	// the impulse split shows up instead of cancelling out.
	a := SpaceObject{X: 496, Y: 500, Radius: defaultRadius, VelocityX: 30, Mass: 80}
	b := SpaceObject{X: 504, Y: 500, Radius: defaultRadius, VelocityX: 0, Mass: 240}

	momentumBefore := a.Mass*a.VelocityX + b.Mass*b.VelocityX
	energyBefore := kineticEnergy(a) + kineticEnergy(b)

	collide(&a, &b, testDistance)

	if a.Mass == 0 || b.Mass == 0 {
		t.Fatal("a hit this hard should bounce, not merge")
	}

	assertClose(t, "momentum", a.Mass*a.VelocityX+b.Mass*b.VelocityX, momentumBefore)

	// They close at 30 and leave at restitution * 30.
	assertClose(t, "separation speed", b.VelocityX-a.VelocityX, restitution*30)

	// The rest of the energy went to heat and sound.
	if energyAfter := kineticEnergy(a) + kineticEnergy(b); energyAfter >= energyBefore {
		t.Errorf("bounce should lose energy: had %v, left with %v", energyBefore, energyAfter)
	}
}

func TestAGentleHitMergesInstead(t *testing.T) {
	a := SpaceObject{X: 496, Y: 500, Radius: defaultRadius, VelocityX: 1, Mass: defaultMass}
	b := SpaceObject{X: 504, Y: 500, Radius: defaultRadius, VelocityX: -1, Mass: defaultMass}

	collide(&a, &b, testDistance)

	if b.Mass != 0 {
		t.Errorf("a touch this slow should merge, got b still at mass %v", b.Mass)
	}
	if a.Mass != 2*defaultMass {
		t.Errorf("want merged mass %v, got %v", 2*defaultMass, a.Mass)
	}
}

func TestObjectsThatCannotGetAwayMerge(t *testing.T) {
	// Already moving apart, but far too slowly to escape each other: gravity is
	// only going to pull them back, so they merge now.
	a := SpaceObject{X: 496, Y: 500, Radius: defaultRadius, VelocityX: -1, Mass: defaultMass}
	b := SpaceObject{X: 504, Y: 500, Radius: defaultRadius, VelocityX: 1, Mass: defaultMass}

	collide(&a, &b, testDistance)

	if b.Mass != 0 {
		t.Error("a pair drifting apart below escape speed should merge")
	}
}

func TestObjectsAlreadyLeavingAreNotHitAgain(t *testing.T) {
	// The state right after a bounce. Every pair is visited once from each
	// object's point of view, so this must be a no-op or the impulse lands
	// twice in a single step.
	a := SpaceObject{X: 496, Y: 500, Radius: defaultRadius, VelocityX: -16, Mass: defaultMass}
	b := SpaceObject{X: 504, Y: 500, Radius: defaultRadius, VelocityX: 16, Mass: defaultMass}

	collide(&a, &b, testDistance)

	if a.Mass == 0 || b.Mass == 0 {
		t.Fatal("a pair leaving this fast should be left alone, not merged")
	}
	assertClose(t, "a velocity", a.VelocityX, -16)
	assertClose(t, "b velocity", b.VelocityX, 16)
}

func TestObjectsInExactlyTheSamePlaceMerge(t *testing.T) {
	// There is no direction to bounce along, and dividing by the distance would
	// turn everything downstream into NaN.
	a := SpaceObject{X: 500, Y: 500, Radius: defaultRadius, Mass: defaultMass}
	b := SpaceObject{X: 500, Y: 500, Radius: defaultRadius, Mass: defaultMass}

	collide(&a, &b, 0)

	if b.Mass != 0 {
		t.Fatal("objects in the same place should merge")
	}
	if math.IsNaN(a.X) || math.IsNaN(a.VelocityX) {
		t.Errorf("merge left NaN behind: %+v", a)
	}
}

// TestABouncingPairEventuallySettles is the whole point of the feature: a hard
// hit bounces, but every bounce is weaker than the last, so the pair works its
// way down to a merge on its own instead of ringing forever.
func TestABouncingPairEventuallySettles(t *testing.T) {
	w := New(10000, 10000)
	w.Objects = []SpaceObject{
		{X: 4996, Y: 5000, Radius: defaultRadius, VelocityX: 20, Mass: defaultMass},
		{X: 5004, Y: 5000, Radius: defaultRadius, VelocityX: -20, Mass: defaultMass},
	}

	w.UpdateSpace()
	if len(w.Objects) != 2 {
		t.Fatal("the first hit should bounce, not merge")
	}

	bounced := false
	for step := 0; step < 5000; step++ {
		w.UpdateSpace()
		if len(w.Objects) == 1 {
			if !bounced {
				t.Fatal("merged without ever separating")
			}
			return
		}
		// They have to actually fly apart, not just sit in the collision band.
		if distanceBetween(w.Objects[0], w.Objects[1]) > 100 {
			bounced = true
		}
	}

	t.Errorf("still %d objects after 5000 steps; the pair never settled", len(w.Objects))
}

// TestObjectsDroppedInSpaceBounceBeforeSettling is the scenario you get by
// clicking twice and letting go: two objects fall together under nothing but
// gravity. They have to visibly bounce off each other a few times on the way to
// merging, which is the whole difference from the old behavior -- tuned too
// timidly, the pair just sticks on first contact and nothing looks different.
func TestObjectsDroppedInSpaceBounceBeforeSettling(t *testing.T) {
	w := New(2000, 2000)
	w.Spawn(900, 1000)
	w.Spawn(1100, 1000)

	bounces := 0
	separating := false

	for step := 0; step < 20000; step++ {
		w.UpdateSpace()

		if len(w.Objects) == 1 {
			if bounces < 2 {
				t.Errorf("want at least 2 bounces before the pair merges, got %d", bounces)
			}
			return
		}

		// A bounce is the moment the gap stops closing and starts opening.
		wasSeparating := separating
		separating = radialSpeed(w.Objects[0], w.Objects[1]) > 0
		if separating && !wasSeparating {
			bounces++
		}
	}

	t.Errorf("the pair never merged; %d bounces so far", bounces)
}

// radialSpeed is how fast the gap between two objects is opening. Negative
// means they are closing on each other.
func radialSpeed(a, b SpaceObject) float64 {
	distance := distanceBetween(a, b)
	if distance == 0 {
		return 0
	}
	return ((b.X-a.X)*(b.VelocityX-a.VelocityX) +
		(b.Y-a.Y)*(b.VelocityY-a.VelocityY)) / distance
}

func kineticEnergy(o SpaceObject) float64 {
	return 0.5 * o.Mass * (o.VelocityX*o.VelocityX + o.VelocityY*o.VelocityY)
}

func distanceBetween(a, b SpaceObject) float64 {
	return math.Hypot(b.X-a.X, b.Y-a.Y)
}

func assertClose(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("want %s %v, got %v", name, want, got)
	}
}
