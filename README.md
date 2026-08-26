## Space Sim

This is (so far) a basic Space Sim.  You run the program and can left-click to add objects to the screen.  These objects will be affected by each other's gravity and will fly around the screen.  Collisions will make them bigger.  Going offscreen makes them appear on the other side.  Space bar freezes time.

Originally written in C++ against raylib, now a from-scratch rewrite in Go.

### Dependencies

Go 1.23+

[Ebitengine](https://ebitengine.org) is the only library, and it arrives through `go.mod` -- there is nothing to install by hand.  On macOS and Windows the build links against system frameworks only; on Linux you'll need the X11 and ALSA development headers.

### Running

```
go mod tidy
go run .
```

`go build` produces a single self-contained binary.

### Layout

| Path | Holds |
| ---- | ----- |
| `world/` | Simulation state and physics.  Knows nothing about drawing. |
| `game/` | Ebitengine glue: input in, pixels out. |
| `main.go` | Window setup and the call into Ebitengine's run loop. |

Ebitengine owns the frame loop, so instead of an explicit draw-begin/draw-end sequence there are `Update` and `Draw` methods on `game.Game`.

#### Dev Todo list


| Feature      | Status |
| ----------- | ----------- |
| Basics of Objects with Gravity | ✅ |
| Collisions      | ✅ |
| Ability to Freeze time and with mouse debug info | ⌛️ |
| Larger map with click to drag | ❌ |
| Customizable config files to set up senarios | ❌ |
| More to come! |  |
