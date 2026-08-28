package world

import (
	"fmt"
	"math"
	"testing"
)

// benchmarkObjects is how many live objects the step benchmarks carry: a busy
// world, but not an unusual one.
const benchmarkObjects = 300

// benchmarkRadius is how far out they sit. Far enough that none of them starts
// inside the star's reach, so what the benchmark measures is the cost of a step
// rather than the cost of a meal.
const benchmarkRadius = 800

// BenchmarkStep measures a single step of a full world, with and without a star,
// and with a star that has already eaten.
//
// A step costs the pair loop -- every live object against every other -- run once
// per substep, so it goes as substeps times objects squared. The object count is
// the same in every case here, which leaves the substep count as the whole story,
// and a star is what drives it: substepCount bounds the strongest pull anything
// could feel, and the star is by far the heaviest thing there is.
//
// The fed cases are the point. A star's mass climbs linearly with what it eats
// while the radius holding gravity at arm's length only climbs as the square root
// of it, so the bound grows without limit and the substep count follows. A fresh
// star is the cheapest a star ever is, and a session only moves away from it.
//
//	go test ./world -bench Step -benchtime 20x
//
// Each case reports its substep count alongside the time, so a change in cost can
// be read straight off as a change in slicing rather than guessed at.
func BenchmarkStep(b *testing.B) {
	cases := []struct {
		name  string
		star  bool
		meals int
	}{
		{name: "NoStar"},
		{name: "FreshStar", star: true},
		{name: "Star300Meals", star: true, meals: 300},
		{name: "Star1000Meals", star: true, meals: 1000},
		{name: "Star3000Meals", star: true, meals: 3000},
	}

	for _, benchmark := range cases {
		b.Run(benchmark.name, func(b *testing.B) {
			for b.Loop() {
				// A fresh world every time: objects are eaten and merged as the
				// step runs, and a benchmark that let that accumulate would be
				// timing a world that empties out from under it.
				b.StopTimer()
				w := benchmarkWorld(benchmark.star, benchmark.meals)
				b.StartTimer()

				w.UpdateSpace()
			}

			b.ReportMetric(float64(benchmarkWorld(benchmark.star, benchmark.meals).Substeps()),
				"substeps")
		})
	}
}

// benchmarkWorld builds a world of benchmarkObjects objects at rest on a ring,
// optionally around a star that has already had the given number of meals.
//
// At rest is deliberate: with nothing moving, fastestSpeed drops out of
// substepCount and the count is entirely the star's doing.
func benchmarkWorld(star bool, meals int) *World {
	w := New(4000, 4000)
	centerX, centerY := w.Width()/2, w.Height()/2

	if star {
		w.SpawnStar(centerX, centerY)
		feedStar(&w.Objects[0], meals)
	}

	for i := 0; i < benchmarkObjects; i++ {
		angle := 2 * math.Pi * float64(i) / benchmarkObjects
		w.Spawn(centerX+benchmarkRadius*math.Cos(angle), centerY+benchmarkRadius*math.Sin(angle))
	}

	return w
}

// feedStar puts the given number of ordinary objects through the star, so a fed
// star's mass and radius are whatever the real merge would have made them rather
// than a formula copied out of absorb and left to drift from it.
func feedStar(star *SpaceObject, meals int) {
	for i := 0; i < meals; i++ {
		meal := SpaceObject{Radius: defaultRadius, Mass: defaultMass}
		star.absorb(&meal)
	}
}

// TestFedStarDrivesSubstepsUp is the benchmark's premise as an assertion, so the
// cost curve is not something only a hand-run benchmark can see. Eating makes a
// star more expensive, monotonically, and there is a point where a step is being
// sliced into dozens of passes over every pair.
func TestFedStarDrivesSubstepsUp(t *testing.T) {
	starless := benchmarkWorld(false, 0).Substeps()

	previous := starless
	for _, meals := range []int{0, 300, 1000, 3000} {
		substeps := benchmarkWorld(true, meals).Substeps()

		t.Logf("%d meals: %d substeps", meals, substeps)
		if substeps <= previous {
			t.Errorf("%d meals: want more substeps than the %d before it, got %d",
				meals, previous, substeps)
		}
		previous = substeps
	}

	// The far end of the table is the part worth pinning: by three thousand meals
	// a step costs an order of magnitude more passes than the same world with no
	// star in it at all.
	fed := benchmarkWorld(true, 3000).Substeps()
	if fed < 10*starless {
		t.Errorf("want a well fed star to cost at least ten times the starless %d substeps, got %d",
			starless, fed)
	}
}

// ExampleWorld_Substeps is here for the numbers rather than the coverage: it puts
// the cost of a fed star in the repository in a form that fails when it changes.
func ExampleWorld_Substeps() {
	for _, meals := range []int{0, 300, 1000, 3000} {
		w := benchmarkWorld(true, meals)
		star := w.Objects[0]
		fmt.Printf("meals %4d: mass %6.0f  radius %4.1f  substeps %d\n",
			meals, star.Mass, star.Radius, w.Substeps())
	}
	// Output:
	// meals    0: mass   5000  radius 30.0  substeps 4
	// meals  300: mass  29000  radius 34.6  substeps 16
	// meals 1000: mass  85000  radius 43.6  substeps 32
	// meals 3000: mass 245000  radius 62.4  substeps 51
}
