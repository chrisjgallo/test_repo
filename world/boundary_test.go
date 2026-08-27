package world

import "testing"

func TestTheWorldWrapsUntilToldOtherwise(t *testing.T) {
	if got := New(1000, 1000).Boundary; got != BoundaryWrap {
		t.Errorf("want a new world to wrap, got %v", got)
	}
}

func TestBoundaryCyclesThroughEveryModeAndBackAround(t *testing.T) {
	w := New(1000, 1000)

	for _, want := range []BoundaryMode{BoundaryWall, BoundaryVoid, BoundaryWrap} {
		w.CycleBoundary()
		if w.Boundary != want {
			t.Fatalf("want %v after cycling, got %v", want, w.Boundary)
		}
	}
}

func TestEveryBoundaryModeHasAName(t *testing.T) {
	names := map[BoundaryMode]string{
		BoundaryWrap: "Wrap",
		BoundaryWall: "Wall",
		BoundaryVoid: "Void",
	}

	for mode := BoundaryMode(0); mode < boundaryModeCount; mode++ {
		want, ok := names[mode]
		if !ok {
			t.Fatalf("mode %d has no expected name; was a mode added?", mode)
		}
		if got := mode.String(); got != want {
			t.Errorf("want name %q, got %q", want, got)
		}
	}
}

func TestAWallSendsObjectsBackInside(t *testing.T) {
	w := New(1000, 1000)
	object := SpaceObject{X: -5, Y: 500, Radius: 1, VelocityX: -10, Mass: defaultMass}

	w.bounceOffEdges(&object)

	// Nudged back to rest against the wall, and turned around minus the share
	// of the speed the impact costs.
	assertClose(t, "x", object.X, object.Radius)
	assertClose(t, "velocity", object.VelocityX, restitution*10)
}

func TestAWallDoesNotTrapObjectsAlreadyLeaving(t *testing.T) {
	// Overlapping the wall but already heading away from it. Reversing here
	// would pin the object to the edge instead of letting it go.
	w := New(1000, 1000)
	object := SpaceObject{X: 0.5, Y: 500, Radius: 1, VelocityX: 5, Mass: defaultMass}

	w.bounceOffEdges(&object)

	assertClose(t, "velocity", object.VelocityX, 5)
}

func TestEveryWallIsSolid(t *testing.T) {
	w := New(1000, 1000)

	for _, wall := range []struct {
		name       string
		object     SpaceObject
		wantX      float64
		wantY      float64
		wantXSpeed float64
		wantYSpeed float64
	}{
		{"left", SpaceObject{X: -5, Y: 500, VelocityX: -10}, 1, 500, restitution * 10, 0},
		{"right", SpaceObject{X: 1005, Y: 500, VelocityX: 10}, 999, 500, -restitution * 10, 0},
		{"top", SpaceObject{X: 500, Y: -5, VelocityY: -10}, 500, 1, 0, restitution * 10},
		{"bottom", SpaceObject{X: 500, Y: 1005, VelocityY: 10}, 500, 999, 0, -restitution * 10},
	} {
		object := wall.object
		object.Radius, object.Mass = 1, defaultMass

		w.bounceOffEdges(&object)

		assertClose(t, wall.name+" wall x", object.X, wall.wantX)
		assertClose(t, wall.name+" wall y", object.Y, wall.wantY)
		assertClose(t, wall.name+" wall x velocity", object.VelocityX, wall.wantXSpeed)
		assertClose(t, wall.name+" wall y velocity", object.VelocityY, wall.wantYSpeed)
	}
}

func TestTheVoidSwallowsWhateverLeaves(t *testing.T) {
	w := New(1000, 1000)
	w.Boundary = BoundaryVoid
	w.Objects = []SpaceObject{
		{X: 995, Y: 500, Radius: 1, VelocityX: 20, Mass: defaultMass},
	}

	// One step carries it out, the next sweeps it up.
	w.UpdateSpace()
	w.UpdateSpace()
	w.UpdateSpace()

	if len(w.Objects) != 0 {
		t.Errorf("want the object gone, got %+v", w.Objects)
	}
}

func TestTheVoidLeavesObjectsInsideAlone(t *testing.T) {
	w := New(1000, 1000)
	w.Boundary = BoundaryVoid
	w.Spawn(500, 500)

	for i := 0; i < 10; i++ {
		w.UpdateSpace()
	}

	if len(w.Objects) != 1 {
		t.Errorf("want the object left alone, got %d objects", len(w.Objects))
	}
}

func TestObjectsStillWrapInWrapMode(t *testing.T) {
	// The original behavior has to survive being made one option among three.
	w := New(1000, 1000)
	w.Boundary = BoundaryWrap
	w.Spawn(1000+roomForError+1, 500)

	w.UpdateSpace()

	if got := w.Objects[0].X; got != roomForError {
		t.Errorf("want object wrapped to x=%v, got %v", float64(roomForError), got)
	}
}
