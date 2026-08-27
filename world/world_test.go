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

func TestSpawnMovingKeepsInitialVelocity(t *testing.T) {
	w := New(1000, 1000)
	w.SpawnMoving(500, 500, 3, -2)

	object := w.Objects[0]
	if object.VelocityX != 3 || object.VelocityY != -2 {
		t.Errorf("want velocity (3, -2), got (%v, %v)", object.VelocityX, object.VelocityY)
	}

	// A lone object drifts along that velocity with nothing to pull on it.
	w.UpdateSpace()

	if got := w.Objects[0].X; got != 503 {
		t.Errorf("want x 503 after one step, got %v", got)
	}
	if got := w.Objects[0].Y; got != 498 {
		t.Errorf("want y 498 after one step, got %v", got)
	}
}

func TestTotalMassSurvivesACollision(t *testing.T) {
	w := New(1000, 1000)
	w.Spawn(500, 500)
	w.Spawn(505, 500)
	w.Spawn(100, 100)

	if got := w.TotalMass(); got != 3*defaultMass {
		t.Errorf("want starting mass %v, got %v", 3*defaultMass, got)
	}

	w.UpdateSpace() // the close pair merges
	w.UpdateSpace() // the absorbed object is swept up

	if len(w.Objects) != 2 {
		t.Fatalf("want 2 objects after the merge, got %d", len(w.Objects))
	}
	if got := w.TotalMass(); got != 3*defaultMass {
		t.Errorf("merging should conserve mass, want %v, got %v", 3*defaultMass, got)
	}
}
