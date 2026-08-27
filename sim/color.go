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

// starColor fills a fixed star. It sits off the end of the ramp rather than at
// the top of it: a star's size says nothing about how much it has eaten, and it
// should read as its own kind of thing rather than as the largest blob on
// screen. Near-white with a warm cast, so it stays the brightest thing out
// there however red the giants around it get.
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
