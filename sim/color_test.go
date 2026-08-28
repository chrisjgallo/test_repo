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

func TestTheGlowSitsBetweenTheStarAndItsOuterEdge(t *testing.T) {
	const radius = 30.0

	layers := glowLayers(radius)
	if len(layers) != glowLayerCount {
		t.Fatalf("want %d layers, got %d", glowLayerCount, len(layers))
	}

	for i, layer := range layers {
		// Anything at or inside the star's own radius is drawn over by the star,
		// and anything past the outer edge is a halo wider than advertised --
		// which the off-screen check is sized against.
		if layer.radius <= radius || layer.radius > radius*glowRadiusScale {
			t.Errorf("layer %d: radius %v is outside (%v, %v]",
				i, layer.radius, radius, radius*glowRadiusScale)
		}
	}
}

func TestTheGlowIsDrawnOutermostFirst(t *testing.T) {
	// The layers stack their alpha up as they are drawn, so the order is what
	// makes the halo brightest against the star and faintest at its edge.
	layers := glowLayers(30)

	for i := 1; i < len(layers); i++ {
		if layers[i].radius >= layers[i-1].radius {
			t.Errorf("layer %d (%v) should be inside layer %d (%v)",
				i, layers[i].radius, i-1, layers[i-1].radius)
		}
	}
}

func TestTheGlowStaysAGlow(t *testing.T) {
	// Stacked, the layers have to come out translucent: a halo that reaches
	// opaque is a bigger star with a hard edge, not light coming off one.
	clear := 1.0
	for _, layer := range glowLayers(30) {
		if layer.color.A == 0 || layer.color.A == 255 {
			t.Fatalf("want a translucent layer, got alpha %d", layer.color.A)
		}
		clear *= 1 - float64(layer.color.A)/255
	}

	if opacity := 1 - clear; opacity > 0.5 {
		t.Errorf("want the halo under half opaque where it meets the star, got %.2f", opacity)
	}
}

func TestTheGlowZoomsWithTheStar(t *testing.T) {
	// The halo is built from the drawn radius, so it tracks the zoom rather than
	// hanging at a fixed size around a star that has grown or shrunk on screen.
	near, far := glowLayers(60), glowLayers(30)

	for i := range far {
		if near[i].radius != 2*far[i].radius {
			t.Errorf("layer %d: want %v at twice the radius, got %v",
				i, 2*far[i].radius, near[i].radius)
		}
	}
}

func TestTheGlowIsTheStarsOwnColor(t *testing.T) {
	// A halo in any other hue reads as a separate object around the star.
	for i, layer := range glowLayers(30) {
		if layer.color.R != starColor.R || layer.color.G != starColor.G ||
			layer.color.B != starColor.B {
			t.Errorf("layer %d: want the star's color %v, got %v", i, starColor, layer.color)
		}
	}
}
