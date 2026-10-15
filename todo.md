# ebitter TODO

Решения по дизайну зафиксированы: движок «all-around», без строгих имплементаций там, где можно
дать гибкость; глобальные дефолты сохраняем ради обратной совместимости.

Все правки — обратносовместимые, если не помечено иначе. После каждой фазы: `gofmt` + `go build ./...`
+ `go vet ./...` + `go test ./...`.

---

## Готово

- [x] Pitch variation в `sound.PlayVariable` (ресемплинг PCM, т.к. в Ebitengine v2.9.9 нет playback rate).
      - `sound/sound.go`: `PlayVariable`, `SetPitchRange`, общий `play`, `resamplePCM`.
      - `sound/sound_test.go`: тесты ресемплера.
      - `text/text.go`, `text/dialog.go`: опциональный разброс pitch для typing-звука (`CharPitchMin/Max`, `CharSoundVolume`).
      - `cmd/demo/dialogs.go`: мягкий разброс `0.85..1.2`.

---

## Фаза 1 — Инъекция зависимостей вместо глобалов ✅

- [x] `events`: инъекция шины.
  - [x] `InteractionZone.Bus *events.SignalBus` (nil → `MasterSignalBus`).
  - [x] `Entity.Bus *events.SignalBus` для проверки коллизий (nil → `MasterSignalBus`).
- [x] `render`: метод `AtlasManager.AnimatedIcon(dir, state)` + `NewAnimatedIconFromPath` делегирует.
  - [x] Удалить мёртвое поле `AtlasManager.BasePath`.
- [x] `fs.FS` как поле (nil → `assetfs.FS`) в `render.AtlasManager`, `tiled.TilesetCache`, `sound.SoundPlayer`.
  - [x] `tiled` читает карту/тайлсеты через `TilesetCache.FS`.

## Фаза 2 — Гибкие конвенции tiled ✅

- [x] `tiled.Schema` со всеми зашитыми именами (ключевые слова групп, имена свойств, дефолтный сигнал, значения триггера, `min_zoom`/`max_zoom`, `blueprint`).
- [x] `LoadMapScaledWith` / `LoadMapSpecScaledWith`; существующие `Load*` используют `DefaultSchema()`.
- [x] Хранить `schema` в `GameMap`; `MinZoom`/`MaxZoom`/`Spawns` через него.

## Фаза 3 — Баги/чистка render ✅

- [x] Направления анимаций: добавлены `AnimationModeReverse` и `AnimationModePingPongReverse`.
- [x] Отвязать поиск иконки от имени директории (учёт `dirIcons`).
- [x] Удалён неиспользуемый `animationModeToString`; корректная отметка директорий как обработанных.
- [x] Тесты анимаций (`render/animation_test.go`).

## Фаза 4 — Конфигурируемость камеры/текста/рендера ✅

- [x] `engine.Camera`: `SetEasing(...)`, `SetZoomFilter(...)`.
- [x] `text`: настраиваемые `FontRoot` и `Filter`.
- [x] `scene.SceneManager`: `FadeColor`.
- [x] **Не-цель:** маркеры парсера `$p/$c/$s/$n/$e/$f` остаются хардкод.

## Фаза 5 — Звук ✅

- [x] **Только OGG**: `RegisterNewSound` декодирует через `audio/vorbis`.
- [x] Контент: `media/sound/*.wav` → `.ogg` (ffmpeg), путь в `cmd/demo/dialogs.go` обновлён.
- [ ] (опц.) один фоновый reaper вместо горутины на каждый звук.
- [ ] (опц.) каналы для BGM (сейчас один слот `bgm`).

---

## Сущности: data-driven система компонентов (core v1)

### Архитектура
- Логика — **в самих компонентах** (для расширяемости), не во внешних системах.
- Всё через **реестр** компонентов.
- `Entity` — контейнер: `ID`, `Layer`, `Tags`, `State map[string]any`, список компонентов,
  отсортированный по `order`, + индекс по имени. У каждого компонента обратная ссылка на владельца.
- Интерфейсы:
  ```go
  type Component interface{ Type() string }
  type Updater       interface{ Update(e *Entity, dt time.Duration) }
  type Drawer        interface{ Draw(e *Entity, s *ebiten.Image) }
  type SignalHandler interface{ OnSignal(e *Entity, sig events.Signal) }
  ```
- `order` в yml **необязателен**: дефолт задаётся при регистрации компонента в реестре, поле в yml переопределяет.
  Сортировка стабильная — при равных order сохраняется порядок добавления.
- Рефлексия только при загрузке yml, в кадре — никакой.
- Программное создание тоже работает: `e.Add(&components.Velocity{X: 1})`.

### Дефолтные order (задокументировать)
```
body         -20   # силы до интеграции
velocity     -10   # интеграция в transform
transform      0
animator      10   # переключает sprite до его отрисовки
sprite        20
hitbox        30
trigger       40
interactable  50
lifetime     100   # последним
```

### Компоненты v1
- [x] `transform` — x, y, scaleX/scaleY, rotation (переиспользует существующий `Transform`).
- [x] `velocity` — x, y (переиспользует `Velocity`); интегрируется в `transform`.
- [x] `body` — gravity, drag, maxSpeed, accel; силы → `velocity`, кламп, grounded.
- [x] `sprite` — icon, state, anchor, offset, flipH/V; рисует через `render`, корень настраиваемый (`components.SpriteRoot`).
- [x] `animator` — default, `on: {signal→state}`; слушает сигналы, переключает `sprite.state`.
- [x] `hitbox` — w/h, offset, layer, mask, solid; SAT через существующий `Collider`.
- [x] `trigger` — w/h, offset, signal, mode(touch/button), once; overlap → эмит сигнала.
- [x] `interactable` — зона, prompt, signal; кнопка внутри зоны (отдельно от `trigger`).
- [x] `lifetime` — ttl; авто-destroy (0 = бесконечно).
- Метаданные сущности (не компоненты): `tags`, `state`, `layer`.

### Пример `media/entities/froggit.yml`
```yaml
id: froggit
layer: entity
tags: [enemy]
state: { met_before: false }
components:
  transform: { x: 0, y: 0, scale: 2 }
  sprite:    { icon: froggit, state: idle, anchor: [0.5, 1.0] }
  animator:  { default: idle, on: { froggit_move: walk, froggit_stop: idle } }
  velocity:  { x: 0, y: 0 }
  body:      { gravity: 0, drag: 0.1, maxSpeed: 120 }
  hitbox:    { w: 42, h: 24, layer: enemy, mask: [player, wall] }
  health:    { max: 20, invuln: 0.5 }
  trigger:   { w: 48, h: 32, signal: talk_froggit, mode: button }
  lifetime:  { ttl: 0 }
```

### Загрузка
- [x] Префабы `media/entities/*.yml` → `Blueprint` (`LoadBlueprint`, `BlueprintLibrary`).
- [x] `Spawn(id, x, y)` с оверрайдами.
- [x] Зависимость `gopkg.in/yaml.v3` (прямая).

### Плейсмент из Tiled (вариант B) — ПОЗЖЕ, после ядра
- [ ] Объекты с строковым свойством `blueprint: froggit` → запросы спавна из `LoadMap`.
- Пока **вариант A**: спавн только через Go/`Spawn`.

### Порядок работ
1. [x] Реестр + контейнер `Entity` + order-сортировка + обратная ссылка на владельца.
2. [x] Порт `transform`, `velocity`, `sprite`, `hitbox`; затем `body`, `animator`, `trigger`, `interactable`, `health`, `lifetime`.
3. [x] YAML-лоадер префабов + `Spawn`.
4. [x] (позже) Tiled-плейсмент (B).
5. [x] Тесты: реестр, порядок обновления, загрузка yml, базовые компоненты.

### Ломающие изменения
- `components.Entity` меняет форму (сейчас не используется: демка на `tiled.ColliderBox`, тесты — на `Collider`).
- `Collider`, `Transform`, `Velocity`, `Sprite` переиспользуются как есть.
- `DrawSpriteOnCollider` — сохранить/перенести.
