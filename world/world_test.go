package world

import (
	"math"
	"testing"
)

func TestGravityPullsObjectsTogether(t *testing.T) {
	w := New(1000, 1000)
	w.Spawn(100, 100)
	w.Spawn(500, 100)

	w.UpdateSpace()

	if w.Objects[0].VelocityX <= 0 {
		t.Errorf("left object should accelerate right, got VelocityX %v", w.Objects[0].VelocityX)
	}
	if w.Objects[1].VelocityX >= 0 {
		t.Errorf("right object should accelerate left, got VelocityX %v", w.Objects[1].VelocityX)
	}
}

func TestCollisionMergesObjects(t *testing.T) {
	w := New(1000, 1000)
	w.Spawn(500, 500)
	w.Spawn(505, 500)

	w.UpdateSpace() // merges, leaving the absorbed object at zero mass
	w.UpdateSpace() // sweeps it up

	if len(w.Objects) != 1 {
		t.Fatalf("want 1 object after collision, got %d", len(w.Objects))
	}
	if got := w.Objects[0].Mass; got != 2*defaultMass {
		t.Errorf("want merged mass %v, got %v", 2*defaultMass, got)
	}
	// Area is conserved, so the merged radius grows by sqrt(2).
	want := math.Sqrt(2) * defaultRadius
	if got := w.Objects[0].Radius; math.Abs(got-want) > 1e-9 {
		t.Errorf("want merged radius %v, got %v", want, got)
	}
}

func TestObjectsWrapAroundEdges(t *testing.T) {
	w := New(1000, 1000)
	w.Spawn(1000+roomForError+1, 500)

	w.UpdateSpace()

	if got := w.Objects[0].X; got != roomForError {
		t.Errorf("want object wrapped to x=%v, got %v", float64(roomForError), got)
	}
}

func TestPausedWorldDoesNotMove(t *testing.T) {
	w := New(1000, 1000)
	w.Spawn(100, 100)
	w.Spawn(500, 100)
	w.TogglePause()

	w.UpdateSpace()

	for i, object := range w.Objects {
		if object.VelocityX != 0 || object.VelocityY != 0 {
			t.Errorf("object %d moved while paused: %+v", i, object)
		}
	}
}
