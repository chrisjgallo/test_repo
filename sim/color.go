package sim

import "image/color"

// Objects are colored by how big they have grown, so a glance at the screen
// tells you which bodies have been eating the others. The ramp runs cool to
// hot: new objects are small and blue, and the giants that have swallowed
// hundreds of them burn red.
//
// The stops are spaced by radius rather than mass because merging conserves
// area, so radius grows as the square root of the number of objects absorbed.
// A radius of 3 is roughly nine objects, 10 is a hundred, 16 is a few hundred:
// close enough to even spacing across a run to keep the whole ramp in use.
var colorRamp = []colorStop{
	{radius: 1, color: color.RGBA{R: 120, G: 170, B: 240, A: 255}},
	{radius: 3, color: color.RGBA{R: 255, G: 255, B: 255, A: 255}},
	{radius: 6, color: color.RGBA{R: 255, G: 228, B: 130, A: 255}},
	{radius: 10, color: color.RGBA{R: 255, G: 155, B: 65, A: 255}},
	{radius: 16, color: color.RGBA{R: 235, G: 75, B: 50, A: 255}},
}

// starColor fills a fixed star, which is off the ramp rather than at the top of
// it. A star's radius does grow as it eats -- absorb conserves area for a fixed
// object as much as for any other, so thirty becomes about thirty-five after
// three hundred meals -- but it is born past the far end of the ramp and only
// climbs from there. The ramp means "this is how much I have eaten", and read
// through it a star would be pinned at the hottest color from its first tick,
// reporting hundreds of meals it has not had and then never moving again.
//
// It is not what tells a star from a big merge, though -- size is, at thirty
// against the handful a merge reaches. The ramp sweeps a continuous path from
// blue through white to red, so every near-white lies close to some point on it;
// this one is a couple of shades off the color of a radius-4 object. Warm, on the
// grounds that the thing in the middle of a solar system is a sun.
var starColor = color.RGBA{R: 255, G: 241, B: 200, A: 255}

// colorStop pins a color to a radius. Anything between two stops is mixed from
// the pair, and anything past the ends takes the end's color unchanged.
type colorStop struct {
	radius float64
	color  color.RGBA
}

// colorForRadius picks the color an object of the given radius is drawn in.
func colorForRadius(radius float64) color.RGBA {
	if radius <= colorRamp[0].radius {
		return colorRamp[0].color
	}

	last := colorRamp[len(colorRamp)-1]
	if radius >= last.radius {
		return last.color
	}

	for i := 1; i < len(colorRamp); i++ {
		low, high := colorRamp[i-1], colorRamp[i]
		if radius > high.radius {
			continue
		}

		progress := (radius - low.radius) / (high.radius - low.radius)
		return blend(low.color, high.color, progress)
	}

	// Unreachable: the checks above cover everything below the last stop.
	return last.color
}

// blend mixes two colors, with progress running from 0 (all of from) to 1 (all
// of to).
func blend(from, to color.RGBA, progress float64) color.RGBA {
	return color.RGBA{
		R: mix(from.R, to.R, progress),
		G: mix(from.G, to.G, progress),
		B: mix(from.B, to.B, progress),
		A: mix(from.A, to.A, progress),
	}
}

func mix(from, to uint8, progress float64) uint8 {
	return uint8(float64(from) + (float64(to)-float64(from))*progress)
}

const (
	// glowRadiusScale is how far a star's halo reaches, as a multiple of the
	// star's own radius. Three is wide enough to read as light coming off the
	// thing rather than as a fatter star, and still narrow enough that an object
	// falling in is visible against it for most of the way down.
	glowRadiusScale = 3.0

	// glowLayerCount is how many circles the halo is built from. Each one is a
	// hard edge, so too few show up as visible rings; six is past the point
	// where the banding is noticeable at the sizes a star is drawn at.
	glowLayerCount = 6

	// glowLayerAlpha is how opaque a single layer is. They stack, so the halo
	// ends up around a third opaque where it meets the star and one layer's
	// worth of nothing at its outer edge -- a light glow, which is the point:
	// the star should still be plainly the brightest thing on the screen.
	glowLayerAlpha = 24
)

// glowColor is the halo's color: the star's own, thinned right down. Anything
// cooler would read as a separate object hanging around the star rather than as
// its light.
//
// It is NRGBA rather than RGBA because the drawing code takes colors as
// premultiplied, which is what RGBA means in this package's terms and would
// need the channels scaled down by hand to match the alpha. NRGBA converts on
// the way through instead, so the channels here are the color as written.
var glowColor = color.NRGBA{R: starColor.R, G: starColor.G, B: starColor.B, A: glowLayerAlpha}

// glowLayer is one circle of a star's halo.
type glowLayer struct {
	radius float64
	color  color.NRGBA
}

// glowLayers is the halo to draw around a star of the given drawn radius,
// outermost first, so that drawing them in order stacks the alpha up towards
// the middle and the halo fades outwards on its own.
//
// The innermost layer stops short of the star rather than sitting on it: the
// star is drawn opaque on top, so a layer under it would be hidden and the
// falloff would start a step in from the edge.
func glowLayers(radius float64) []glowLayer {
	step := (glowRadiusScale - 1) / glowLayerCount

	layers := make([]glowLayer, 0, glowLayerCount)
	for i := 0; i < glowLayerCount; i++ {
		layers = append(layers, glowLayer{
			radius: radius * (glowRadiusScale - float64(i)*step),
			color:  glowColor,
		})
	}

	return layers
}
