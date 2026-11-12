<h1><img valign="middle" width="50" height="50" alt="bitter" src="https://github.com/user-attachments/assets/42f58e98-4ae7-43d6-9822-08e848eb1b46"/> ebitter</h1>

A set of utilities for [ebitngine](https://ebitengine.org) to make writing games in Go easier!

All of the tooling required to work with Ebitter is free and opensource.

<img width="1920" height="1080" alt="image" src="https://github.com/user-attachments/assets/1d4248ed-79ef-4a70-8b55-ac551697bc91" />
<sub>A demo scene running in Ebitter showcasing tilemap loading; <a href="https://github.com/aseprite/aseprite/">Aseprite</a> and <a href="https://github.com/mapeditor/tiled/">Tiled</a> are open. Tilemap and all the artwork are from UNDERTALE by Toby Fox and are used for demonstration purposes.</sub>
<br></br>

Ebitter offers you:
- scene management;
- camera movement;
- [Tiled](https://github.com/mapeditor/tiled/) map loading;
- [Aseprite](https://github.com/aseprite/aseprite/) sprite-sheet loading and BYOND-like stateful icons;
- a component system, YAML prototypes;
- tweening;
- signal buses;
- update/draw/delete queues and auto-registration;
- custom font rendering;
- a dialogue scripting system;
- font-atlas generation;
- OGG sound playback;
- an asset filesystem for embedding!

Check out the [quickstart guide](./QUICKSTART.md) to see how to use Ebitter or run the [web demo](http://mcbalaam.github.io/ebitter)!

## Run and test

```sh
go run ./cmd/demo
go test ./...
```

Other useful checks are `go build ./...` and `go vet ./...`.
