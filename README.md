# ebitter

A set of tools and upgrades for [ebitngine](https://ebitengine.org) to make writing games in Go easier!

Demo here: https://mcblm.xyz/ebitter

What I currently have in store:
- scenes: proper `Scene` management and swapping;
- `Camera`: move, tilt and zoom the camera around the scene;
- sprite atlas cutting: create animated multi-`state` sprites using Aseprite's tag system and sprite sheet exports;
- font cutting: use any `.ttf` font in your game!
- UTDR-like dialogue scripting system: a proper `DialogueHandler` and markers to pause/end/otherwise modify the text;
- separate `Update`, `Delete` and layered `Draw` queues;
- componentable `Object` to be used as a base for any on-screen things;
- SAT collision detection, hitboxes as a component;
- a signal bus and a bare event system to subscribe and listen to events;
- a simplified input handler with frame-perfect snapshots;
- improved sound playing: registering sounds to be reused, pitch variation, volume controls.
