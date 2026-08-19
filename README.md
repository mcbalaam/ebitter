# ebitter

A set of tools and upgrades for [ebitngine](https://ebitengine.org) to make writing games in Go easier!

All of the tooling required to work with Ebitter is free and opensource.

<img width="1920" height="1080" alt="image" src="https://github.com/user-attachments/assets/d39a8179-87c3-48bb-abea-a67b629caccd" />
<br>

<sub>A demo scene running in Ebitter showcasing tilemap loading; [Aseprite](https://github.com/aseprite/aseprite/) and [Tiled](https://github.com/mapeditor/tiled/) are open. Tilemap and all the artwork are from UNDERTALE by Toby Fox and are used for demonstration purposes.</sub>


Live demo here: https://mcblm.xyz/ebitter
<br>
<br>

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
