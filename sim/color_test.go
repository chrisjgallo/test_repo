package sim

import (
	"image/color"
	"testing"
)

func TestSmallObjectsUseTheColdEndOfTheRamp(t *testing.T) {
	coldest := colorRamp[0].color

	// Freshly spawned objects, and anything smaller, are the first stop.
	for _, radius := range []float64{0, 0.5, 1} {
		if got := colorForRadius(radius); got != coldest {
			t.Errorf("radius %v: want %v, got %v", radius, coldest, got)
		}
	}
}

func TestGiantObjectsClampToTheHotEnd(t *testing.T) {
	hottest := colorRamp[len(colorRamp)-1].color

	// Nothing past the last stop keeps getting redder; it just stays there.
	for _, radius := range []float64{16, 40, 1000} {
		if got := colorForRadius(radius); got != hottest {
			t.Errorf("radius %v: want %v, got %v", radius, hottest, got)
		}
	}
}

func TestColorAtAStopIsExactlyThatStop(t *testing.T) {
	for _, stop := range colorRamp {
		if got := colorForRadius(stop.radius); got != stop.color {
			t.Errorf("radius %v: want %v, got %v", stop.radius, stop.color, got)
		}
	}
}

func TestColorBetweenStopsIsMixed(t *testing.T) {
	// Halfway between the first two stops is the average of the two, give or
	// take the truncation to whole bytes.
	want := color.RGBA{R: 187, G: 212, B: 247, A: 255}

	if got := colorForRadius(2); got != want {
		t.Errorf("want %v halfway between the first two stops, got %v", want, got)
	}
}

func TestTheRampHasNoVisibleSeams(t *testing.T) {
	// Objects grow smoothly, so their color should too: no stop should show up
	// as a sudden jump as a body merges its way up the ramp.
	const maxJump = 2

	previous := colorForRadius(0)
	for radius := 0.0; radius <= 20; radius += 0.01 {
		current := colorForRadius(radius)

		for _, channel := range []struct {
			name     string
			from, to uint8
		}{
			{"red", previous.R, current.R},
			{"green", previous.G, current.G},
			{"blue", previous.B, current.B},
		} {
			if difference(channel.from, channel.to) > maxJump {
				t.Fatalf("radius %.2f: %s jumped from %d to %d",
					radius, channel.name, channel.from, channel.to)
			}
		}

		previous = current
	}
}

func difference(a, b uint8) int {
	if a > b {
		return int(a) - int(b)
	}
	return int(b) - int(a)
}
