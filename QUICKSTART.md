# Ebitter Quickstart

A short guide to the core pieces of [ebitter](https://github.com/mcbalaam/ebitter):
scenes, Tiled maps, the camera, Aseprite icons, entities and the signal bus. The
snippets here are distilled from `cmd/demo/`, so read that directory for a
complete working example.

The module path is `github.com/mcbalaam/ebitter`.

## 1. Wiring up the game

Ebitter works on top of [Ebitengine](https://ebitengine.org). Your `Game` struct
drives a `scene.SceneManager` every frame:

```go
package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mcbalaam/ebitter/assetfs"
	"github.com/mcbalaam/ebitter/queues"
	"github.com/mcbalaam/ebitter/scene"
)

func init() {
	assetfs.SetFS(Media) // see "assets" below
}

type Game struct {
	sm *scene.SceneManager
}

func (g *Game) Update() error {
	g.sm.Update(time.Second / 60)
	queues.DefaultDeleteQueue.Execute()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.sm.Draw(screen)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return 640, 480
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("My Game")

	game := &Game{sm: &scene.SceneManager{}}
	game.sm.Push(NewMenuScene())

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
```

### Assets

The `assetfs` package holds the filesystem every loader reads from. On desktop
it defaults to `os.DirFS(".")`. To ship assets with the binary, embed them and
point `assetfs` at the embed:

```go
import (
	"embed"

	"github.com/mcbalaam/ebitter/assetfs"
	"github.com/mcbalaam/ebitter/components"
	"github.com/mcbalaam/ebitter/text"
)

//go:embed media
var Media embed.FS

func init() {
	assetfs.SetFS(Media)
	components.SpriteRoot = "media/sprites"
	text.FontRoot = "media/sprites"
}
```

`Media` is defined by your application, not by ebitter. The library does not
bundle a default game asset directory. If you use blueprint `sprite` components
or the `text` package, configure their asset roots for your own layout as shown.

All paths passed to loaders (Tiled, Aseprite, sound) are relative to this
filesystem, e.g. `media/maps/ruins.tmx`.

## 2. Scenes

A scene implements the `scene.Scene` interface:

```go
type Scene interface {
	Update(dt time.Duration)
	Draw(screen *ebiten.Image)
}
```

```go
type menuScene struct {
	input *engine.Input
}

func (s *menuScene) Update(dt time.Duration) {
	s.input.Update()
	if s.input.JustPressed(ebiten.KeyEnter) {
		// ...
	}
}

func (s *menuScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
}
```

The `SceneManager` is a stack. The top scene is the active one:

| Method | Effect |
| --- | --- |
| `sm.Push(s)` | put `s` on top (used for the first scene) |
| `sm.Pop()` | remove the top scene immediately |
| `sm.SwitchTo(s)` | fade out, replace top, fade in |
| `sm.PopFaded()` | fade out, pop top, fade in |

To reach the manager from inside a scene, store it on the scene (e.g.
`game.sm`), or hold a reference passed to the scene's constructor. When you
need to switch scenes:

```go
sm.SwitchTo(NewPlayScene()) // fade transition to the next scene
```

`GameScene` (`scene.NewGameScene()`) is a convenience scene that runs the
global queues for you — see section 6.

## 3. Loading a Tiled map

```go
import "github.com/mcbalaam/ebitter/tiled"

m, err := tiled.LoadMapScaled("media/maps/ruins.tmx", 2)
```

`LoadMap(path)` is the same with scale `1`. Both parse the `.tmx` and
pre-render the tile layers. If you only need the logical data (colliders,
interactions, points) and want to defer the rendering until the graphics
context is ready, use the "spec" variant instead:

```go
m, err := tiled.LoadMapSpecScaled("media/maps/ruins.tmx", tiled.DefaultTilesetCache, 2)
// later, once inside Update (Ebiten graphics context is running):
if err := m.BuildLayers(tiled.DefaultTilesetCache); err != nil { /* ... */ }
```

Drawing and querying:

```go
// draw all visible layers, offset by a scroll/camera origin (see section 7)
m.DrawLayers(screen, scrollX, scrollY)

// world size, scaled
m.Width()  // int
m.Height() // int

// named point objects (e.g. a "start" spawn marker)
if p, ok := m.Point("start"); ok {
	if transform := player.Transform(); transform != nil {
		transform.SetPosition(p.X, p.Y)
	}
}

// Tiled objects with a string property `blueprint` are returned as requests.
// Create them with the components BlueprintLibrary (see section 5).
for _, spawn := range m.Spawns() {
	entity, err := library.Spawn(spawn.Blueprint, spawn.X, spawn.Y)
	if err != nil {
		return err // e.g. from your scene constructor
	}
	entity.Schedule()
}

// Static Tiled collider boxes and interaction zones are available as:
m.Colliders()
m.Interactions()
```

### Map authoring notes

- Collision objects come from object groups whose name/class contains
  `collision` or `hitbox`, or that have a boolean `collision` property.
- Interaction zones come from groups whose name/class contains `interact`, or
  objects with an `interaction`/`signal` property. Each zone reads a `signal`
  string property (default `"interact"`) and a `trigger` property (`button` or
  `touch`). A `button` zone fires when you press interact inside it; a `touch`
  zone fires when you enter it.
- Point objects (zero-size markers) are collected by name for spawn points.

## 4. Loading an Aseprite icon

Ebitter reads Aseprite's spritesheet export: a `.png` plus its `.json`
(`frameTags` become animation states). The path is the directory holding the
pair:

```go
import "github.com/mcbalaam/ebitter/render"

icon, err := render.NewAnimatedIconFromPath("media/sprites/frisk", "down")
```

`"down"` is the initial state, taken from the Aseprite tag name. Then per frame:

```go
_ = icon.SetIconState("right_walk") // switch animation
icon.Update(dt)                     // advance frames
icon.Draw(screen, x, y, scaleX, scaleY, 0)
```

The path is the directory (relative to the `assetfs` root) holding the
`.png`/`.json` pair; pass it in full as above.

## 5. Component entities and prototypes

An entity is a container for components. Components are registered by name;
built-ins include `transform`, `velocity`, `sprite`, `hitbox`, `body`,
`animator`, `trigger`, `interactable` and `lifetime`. Components may implement
`Updater`, `Drawer`, `SignalHandler` and/or `Initializer`. Update and draw order
comes from registration and can be overridden with `order` in YAML.

Create a YAML prototype in the application's asset filesystem, for example
`media/entities/hero.yml`:

```yaml
id: hero
layer: entity
tags: [player]
components:
  transform: { x: 0, y: 0, scale: 2 }
  sprite:    { icon: hero, state: idle, anchor: [0.5, 1.0] }
  velocity:  { x: 0, y: 0 }
  hitbox:    { w: 16, h: 20, layer: player }
```

Load it through a `BlueprintLibrary`. Loading validates the document and each
registered component's YAML schema; `Spawn` creates a fresh entity and sets its
transform position:

```go
lib := components.NewBlueprintLibrary(nil, "media/entities") // nil uses assetfs.FS
player, err := lib.Spawn("hero", 80, 100)
if err != nil {
	return err
}

player.Schedule() // default update and layered draw queues
```

The application must register custom component types before loading prototypes
that use them. Registration provides a factory and a validator:

```go
import (
	"github.com/mcbalaam/ebitter/components"
	"gopkg.in/yaml.v3"
)

if err := components.Register("myComponent", 15,
	func(node *yaml.Node) (components.Component, error) {
		return &MyComponent{}, nil
	},
	func(node *yaml.Node) error {
		// Validate the YAML shape without creating runtime resources.
		return nil
	},
); err != nil {
	panic(err)
}
```

Implement `Type() string` on the component. The registration name must match
the returned type and the YAML key. Implement `Update`, `Draw` or `OnSignal` as
needed; those methods receive the owning `*components.Entity`, which provides
access to sibling components. `Entity.Add` also supports programmatic creation.

`Entity.Update(dt)`, `Entity.Draw(screen)` and `Entity.OnSignal(signal)` can be
called directly. `Entity.Schedule()` uses the package's default queues, which
are global; custom scenes can run the queues themselves. `Entity.Destroy()`
removes it from the default queues.

## 6. Events (the signal bus)

The package provides `events.MasterSignalBus` as a convenience default. A scene
can instead own a `SignalBus` and assign it to entities/zones that emit signals.
A `Signal` has a `Name`, `Source` and optional `Data`.

```go
import "github.com/mcbalaam/ebitter/events"

// subscribe an entity (or anything) to the default signal bus
events.MasterSignalBus.Subscribe("enemy_died", player, func(s events.Signal) {
	// s.Source is the emitter, s.Data the payload
})

// emit on the default bus
events.MasterSignalBus.Emit("enemy_died", enemy, points)
```

Entity collisions emit automatically: `player.CheckCollisionsWithList(enemies, "enemy_touched")`
emits `"enemy_touched"` with `Source` = player and `Data` = the collided enemy.

Interaction zones from a Tiled map also emit on the bus via `zone.Check(...)` —
subscribe to the zone's `signal` property to react:

```go
events.MasterSignalBus.Subscribe("open_door", player, func(s events.Signal) {
	// zone fired; s.Data is the *tiled.InteractionZone
})
```

## 7. Camera

`engine.Camera` is a world-coordinate camera: `X, Y` is the world point shown
at the center of the screen — the camera's source of truth, kept clamped so it
always *is* the on-screen center — and `Zoom` scales the view. Render the world
inside `Apply`; it takes care of panning, zooming (nearest-neighbour) and
keeping its own viewport in sync:

```go
import "github.com/mcbalaam/ebitter/engine"

cam := engine.NewCamera()
cam.SetBounds(float64(m.Width()), float64(m.Height())) // clamp the center to the map
cam.SetViewport(640, 480)                              // screen size; Apply refreshes it from the target
cam.SetZoomLimits(m.MinZoom(), m.MaxZoom())            // map float properties min_zoom / max_zoom

func (s *playScene) Update(dt time.Duration) {
	// tweens (smooth camera, TextStrings, ...) live on the default update queue;
	// scene.GameScene runs it for you, custom scenes must call it once per frame:
	queues.DefaultUpdateQueue.Execute(dt)

	cam.CenterOn(player.Transform().X, player.Transform().Y)
}

func (s *playScene) Draw(screen *ebiten.Image) {
	screen.Fill(bg)
	cam.Apply(screen, func(target *ebiten.Image, ox, oy float64) {
		m.DrawLayers(target, ox, oy) // world drawing goes in here,
		player.Draw(target)          // offset by ox/oy
	})
	// HUD and dialogs drawn after Apply stay in screen space
}
```

- Zoom snaps by default. `cam.SetZoomLimits(min, max)` restricts it, and maps
  can carry the limits themselves as float properties `min_zoom` / `max_zoom`,
  read via `m.MinZoom()` / `m.MaxZoom()`. If zooming out makes the viewport
  larger than the map, the center pins to the map midpoint.
- `cam.SetSmooth(true)` makes every `CenterOn` and `SetZoom` transition run
  through a tween (durations: `FollowDuration`, `ZoomDuration`). This requires
  `queues.DefaultUpdateQueue` to be executed each frame.
- The demo showcases smooth follow on key `3`.

## Putting it together

1. `init()` — set `assetfs` (embed or working directory).
2. Build a `Game` that calls `sm.Update`/`sm.Draw` and the delete queue.
3. Write scenes implementing `Scene`; push the first, switch later.
4. Load maps with `tiled.LoadMapScaled`, icons with
   `render.NewAnimatedIconFromPath`.
5. Compose an entity from components in code or load it from a YAML prototype;
   schedule it on the default queues or call its lifecycle methods yourself.
6. Decouple logic with `events.MasterSignalBus.Subscribe`/`Emit`.
7. Frame the world with `engine.Camera` (`CenterOn`/`SetZoom`/`Apply`, optional
   smooth follow), and run the update queue each frame for tweens to move.

See `cmd/demo/` for a full example that ties these together.
