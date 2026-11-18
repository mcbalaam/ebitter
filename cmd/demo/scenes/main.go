package scenes

import (
	"fmt"
	"image/color"
	"image/png"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/mcbalaam/ebitter/assetfs"
	democomponents "github.com/mcbalaam/ebitter/cmd/demo/components"
	"github.com/mcbalaam/ebitter/components"
	"github.com/mcbalaam/ebitter/engine"
	"github.com/mcbalaam/ebitter/queues"
	"github.com/mcbalaam/ebitter/render"
	"github.com/mcbalaam/ebitter/scene"
	"github.com/mcbalaam/ebitter/text"
	"github.com/mcbalaam/ebitter/tween"
)

type demoScene struct {
	runtime               *scene.Runtime
	camera                *engine.Camera
	frame                 *ebiten.Image
	preview               *ebiten.Image
	buttons               map[string]*components.Entity
	buttonNames           []string
	selectedButton        int
	menuActive            bool
	menuConfirmed         bool
	selectedMenuOption    menuOption
	menuEntries           []*menuEntry
	menuIndex             int
	menuOptions           map[string][]menuOption
	prototypes            *components.BlueprintLibrary
	battle                *democomponents.BattleController
	actionDialoguePending bool
	commandDialogueTurn   int
	soul                  *render.AnimatedIcon
	phaseLabel            *text.TextString
	playerCombatant       *democomponents.Combatant
	playerNameText        *text.TextString
	playerLevelText       *text.TextString
	playerHPText          *text.TextString
	playerHPValueText     *text.TextString
	playerHPValue         string
	playerHPValueX        float64
	soulX, soulY          float64
	input                 *demoInput
}

type menuEntry struct {
	option  menuOption
	text    *text.TextString
	anchorX float64
	anchorY float64
}

type demoInput struct {
	pressed   map[ebiten.Key]bool
	justPress map[ebiten.Key]bool
}

func newDemoInput() *demoInput {
	return &demoInput{pressed: make(map[ebiten.Key]bool), justPress: make(map[ebiten.Key]bool)}
}

func (i *demoInput) Update() {
	clear(i.justPress)
	for _, key := range inpututil.AppendJustPressedKeys(nil) {
		i.justPress[key] = true
	}
	for _, key := range []ebiten.Key{
		ebiten.KeyArrowLeft, ebiten.KeyArrowRight, ebiten.KeyArrowUp, ebiten.KeyArrowDown,
		ebiten.KeyZ, ebiten.KeyX,
	} {
		i.pressed[key] = ebiten.IsKeyPressed(key)
	}
}

func (i *demoInput) justPressed(key ebiten.Key) bool {
	return i.justPress[key]
}

type BattleBox struct {
	corners [4]tween.Vec2 // top-left, top-right, bottom-right, bottom-left
	stroke  float32
	color   color.Color
	updates *queues.UpdateQueue
	resize  *tween.Tween[[4]tween.Vec2]
}

func NewBattleBox(x, y, width, height float64, updates *queues.UpdateQueue) *BattleBox {
	return &BattleBox{
		corners: [4]tween.Vec2{{X: x, Y: y}, {X: x + width, Y: y}, {X: x + width, Y: y + height}, {X: x, Y: y + height}},
		stroke:  5,
		color:   color.White,
		updates: updates,
	}
}

func (b *BattleBox) Type() string { return "battleBox" }

// Resize tweens all four corners from their current positions to the supplied
// top-left, top-right, bottom-right, and bottom-left positions.
func (b *BattleBox) Resize(topLeft, topRight, bottomRight, bottomLeft tween.Vec2, duration time.Duration) {
	if b.resize != nil {
		tween.Stop(b.resize)
	}
	to := [4]tween.Vec2{topLeft, topRight, bottomRight, bottomLeft}
	if duration <= 0 {
		b.corners = to
		return
	}
	from := b.corners
	lerp := func(a, z [4]tween.Vec2, t float64) [4]tween.Vec2 {
		var out [4]tween.Vec2
		for i := range out {
			out[i] = tween.Vec2{X: a[i].X + (z[i].X-a[i].X)*t, Y: a[i].Y + (z[i].Y-a[i].Y)*t}
		}
		return out
	}
	b.resize = tween.New(from, to, duration, func(corners [4]tween.Vec2) {
		b.corners = corners
	}, lerp).SetEasing(tween.EaseInOutCubic).ScheduleOn(b.updates)
}

func (b *BattleBox) Draw(_ *components.Entity, screen *ebiten.Image) {
	// Draw axis-aligned boxes as four joined filled strips. Unlike a stroked
	// path, this produces square, pixel-aligned joins at every corner.
	if b.corners[0].Y == b.corners[1].Y &&
		b.corners[1].X == b.corners[2].X &&
		b.corners[2].Y == b.corners[3].Y &&
		b.corners[3].X == b.corners[0].X {
		left, top := b.corners[0].X, b.corners[0].Y
		right, bottom := b.corners[2].X, b.corners[2].Y
		width, height := right-left, bottom-top
		stroke := float64(b.stroke)
		vector.DrawFilledRect(screen, float32(left), float32(top), float32(width), float32(stroke), b.color, false)
		vector.DrawFilledRect(screen, float32(left), float32(bottom-stroke), float32(width), float32(stroke), b.color, false)
		vector.DrawFilledRect(screen, float32(left), float32(top+stroke), float32(stroke), float32(height-2*stroke), b.color, false)
		vector.DrawFilledRect(screen, float32(right-stroke), float32(top+stroke), float32(stroke), float32(height-2*stroke), b.color, false)
		return
	}

	var outline vector.Path
	outline.MoveTo(float32(b.corners[0].X), float32(b.corners[0].Y))
	for _, point := range b.corners[1:] {
		outline.LineTo(float32(point.X), float32(point.Y))
	}
	outline.Close()
	vector.StrokePath(screen, &outline,
		&vector.StrokeOptions{Width: b.stroke, LineJoin: vector.LineJoinMiter},
		&vector.DrawPathOptions{AntiAlias: false})
}

func NewDemoScene() (*demoScene, error) {
	s := &demoScene{
		runtime: scene.NewRuntime(), buttons: make(map[string]*components.Entity),
		buttonNames:         []string{"fightButton", "actButton", "itemButton", "mercyButton"},
		input:               newDemoInput(),
		commandDialogueTurn: -1,
	}
	s.camera = engine.NewCamera()
	s.camera.SetZoomLimits(1, 10)
	s.camera.SetEasing(tween.EaseOutExpo)
	s.camera.SetViewport(640, 480)
	s.camera.CenterOn(320, 240)
	s.camera.SetZoom(1.3)
	s.camera.SetUpdateQueue(s.runtime.Queues.Update)
	s.camera.SetSmooth(true)
	s.camera.ZoomDuration(2 * time.Second)
	s.camera.SetZoom(1)
	if err := initDialogAudio(s.runtime.Dialog); err != nil {
		return nil, fmt.Errorf("initialize battle audio: %w", err)
	}
	backgroundSkip := func() bool { return s.dialogueXPressed() }
	s.runtime.Dialog.ShowAllChecker = backgroundSkip
	s.runtime.Dialog.AdvanceChecker = func() bool { return s.dialogueAdvancePressed() }
	previewFile, err := assetfs.FS.Open("media/sprites/preview.png")
	if err != nil {
		return nil, fmt.Errorf("open interface preview: %w", err)
	}
	previewSource, decodeErr := png.Decode(previewFile)
	closeErr := previewFile.Close()
	if decodeErr != nil {
		return nil, fmt.Errorf("decode interface preview: %w", decodeErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close interface preview: %w", closeErr)
	}
	s.preview = ebiten.NewImageFromImage(previewSource)
	s.preview.Clear()
	s.soul, err = render.NewAnimatedIconFromPath("media/sprites/soul", "idle")
	if err != nil {
		return nil, fmt.Errorf("load soul sprite: %w", err)
	}
	s.selectedButton = 0

	box := components.NewEntity("battleBox")
	box.Layer = queues.LayerArena
	box.Queues = s.runtime.Queues
	box.Add(NewBattleBox(32, 250, 576, 140, s.runtime.Queues.Update))
	box.Schedule()

	buttonLibrary := components.NewBlueprintLibrary(nil, "prototypes")
	s.prototypes = buttonLibrary
	buttons := []string{"actButton", "fightButton", "itemButton", "mercyButton"}
	for _, button := range buttons {
		blueprint, err := buttonLibrary.Load(button)
		if err != nil {
			return nil, fmt.Errorf("load %s prototype: %w", button, err)
		}
		entity, err := blueprint.InstantiateWith(s.runtime.Signals, s.runtime.Queues)
		if err != nil {
			return nil, fmt.Errorf("instantiate %s: %w", button, err)
		}
		entity.Schedule()
		s.buttons[button] = entity
	}
	encounterBlueprint, err := buttonLibrary.Load("trainingEncounter")
	if err != nil {
		return nil, fmt.Errorf("load training encounter prototype: %w", err)
	}
	encounter, err := encounterBlueprint.Instantiate()
	if err != nil {
		return nil, fmt.Errorf("instantiate training encounter: %w", err)
	}
	battle, ok := components.ComponentOf[*democomponents.BattleController](encounter, "battleController")
	if !ok {
		return nil, fmt.Errorf("training encounter has no battleController")
	}
	inventory, _ := components.ComponentOf[*democomponents.Inventory](encounter, "inventory")
	playerBlueprint, err := buttonLibrary.Load(battle.Player)
	if err != nil {
		return nil, fmt.Errorf("load player prototype: %w", err)
	}
	playerEntity, err := playerBlueprint.Instantiate()
	if err != nil {
		return nil, fmt.Errorf("instantiate player prototype: %w", err)
	}
	playerCombatant, ok := components.ComponentOf[*democomponents.Combatant](playerEntity, "combatant")
	if !ok {
		return nil, fmt.Errorf("player prototype %q has no combatant", battle.Player)
	}
	s.playerCombatant = playerCombatant
	playerName := playerCombatant.Name
	if playerName == "" {
		playerName = battle.Player
	}
	statusStyle := text.TextStyle{
		FontName: "mars", ScaleX: 0.25, ScaleY: 0.25, FontHeight: 24, Color: color.White,
	}
	marsFont, err := render.NewAnimatedIconFromPath("media/sprites/mars", " ")
	if err != nil {
		return nil, fmt.Errorf("load Mars font for player status: %w", err)
	}
	levelLabel := fmt.Sprintf("LV%d", playerCombatant.Level)
	levelX := 30 + statusTextWidth(marsFont, playerName, statusStyle.ScaleX) + 24
	hpX := levelX + statusTextWidth(marsFont, levelLabel, statusStyle.ScaleX) + 24
	s.playerNameText = text.NewTextString(playerName, 30, 384, statusStyle)
	s.playerLevelText = text.NewTextString(levelLabel, levelX, 384, statusStyle)
	s.playerHPText = text.NewTextString("HP", hpX, 384, statusStyle)
	s.playerHPValueX = hpX + float64(hpBarWidth(playerCombatant.MaxHP)) + 12
	s.playerHPValueText = text.NewTextString("", s.playerHPValueX, 384, statusStyle)

	s.battle = battle
	s.menuOptions, err = buildMenuOptions(buttonLibrary, battle, inventory)
	if err != nil {
		return nil, fmt.Errorf("build battle menus: %w", err)
	}
	s.phaseLabel = text.NewTextString(string(battle.Phase), 8, -8, text.TextStyle{
		FontName: "mars", ScaleX: 0.2, ScaleY: 0.2, FontHeight: 20, Color: color.White,
	})
	s.syncButtonSelection()
	return s, nil
}

// SetButtonActive changes one of the prototype-created buttons via its signal.
// Names are actButton, fightButton, itemButton, and mercyButton.
func (s *demoScene) SetButtonActive(name string, active bool) error {
	button, ok := s.buttons[name]
	if !ok {
		return fmt.Errorf("battle button %q not found", name)
	}
	democomponents.SetBattleButtonActive(button, active)
	return nil
}

func (s *demoScene) Runtime() *scene.Runtime { return s.runtime }
func (s *demoScene) Pause()                  { s.runtime.Queues.Pause() }
func (s *demoScene) Resume()                 { s.runtime.Queues.Resume() }

func (s *demoScene) Update(dt time.Duration) {
	s.input.Update()
	s.updatePhaseLabel()
	s.maybeShowCommandDialogue()
	if s.runtime.Dialog.Active() {
		s.runtime.Dialog.Update(dt)
		if !s.runtime.Dialog.Active() && s.actionDialoguePending {
			s.finishPlayerActionDialogue()
		}
	}
	s.updateSoul(dt)
	s.runtime.Dialog.SetVisible(!s.menuActive)
	s.runtime.Queues.ExecuteUpdate(dt)
	s.runtime.Queues.ExecuteDelete()
}

func (s *demoScene) updatePhaseLabel() {
	if s.battle != nil && s.phaseLabel != nil {
		phase := string(s.battle.Phase)
		if s.phaseLabel.Text != phase {
			s.phaseLabel.SetText(phase)
		}
	}
}

func (s *demoScene) dialogueXPressed() bool {
	return !s.menuActive && s.input.justPressed(ebiten.KeyX)
}

func (s *demoScene) dialogueAdvancePressed() bool {
	return !s.menuActive &&
		s.battle != nil &&
		s.battle.Phase == democomponents.BattlePhasePlayerAction &&
		s.input.justPressed(ebiten.KeyZ)
}

func commandLineForTurn(lines []string, turn int) (string, bool) {
	if turn < 0 || turn >= len(lines) {
		return "", false
	}
	return lines[turn], true
}

func (s *demoScene) maybeShowCommandDialogue() {
	if s.battle == nil || s.battle.Phase != democomponents.BattlePhaseCommandSelect ||
		s.commandDialogueTurn == s.battle.Turn || s.runtime.Dialog.Active() {
		return
	}
	s.commandDialogueTurn = s.battle.Turn
	s.runtime.Dialog.SetVisible(true)
	line, ok := commandLineForTurn(s.battle.CommandLines, s.battle.Turn)
	if ok {
		s.runtime.Dialog.Show(line, battleDialogStyle(), sndText)
	}
}

func (s *demoScene) soulVisible() bool {
	if s.battle == nil {
		return s.soul != nil
	}
	return s.soul != nil && (s.battle.Phase == democomponents.BattlePhaseCommandSelect ||
		s.battle.Phase == democomponents.BattlePhaseTargetSelect)
}

func (s *demoScene) openMenu() {
	if s.selectedButton < 0 || s.selectedButton >= len(s.buttonNames) {
		return
	}
	options := s.menuOptions[s.buttonNames[s.selectedButton]]
	s.closeMenu()
	s.menuIndex = 0
	style := text.TextStyle{
		FontName: "determination", ScaleX: 0.25, ScaleY: 0.25,
		FontHeight: 130, Color: color.White,
		CharSpacing: 4,
	}
	if len(options) == 0 {
		options = []menuOption{{label: "(No items)"}}
	}
	for i, option := range options {
		y := 287.0 + float64(i)*32
		entry := &menuEntry{
			option:  option,
			anchorX: 72, anchorY: y,
			text: text.NewTextString("* "+option.label, 90, y-27, style),
		}
		entry.text.Queues = s.runtime.Queues
		entry.text.Show()
		s.menuEntries = append(s.menuEntries, entry)
	}
	s.menuActive = len(s.menuEntries) > 0
	if s.menuActive {
		s.moveSoulToMenuEntry()
	}
}

func (s *demoScene) closeMenu() {
	for _, entry := range s.menuEntries {
		entry.text.Destroy()
	}
	s.menuEntries = nil
}

func (s *demoScene) moveSoulToMenuEntry() {
	if s.menuIndex < 0 || s.menuIndex >= len(s.menuEntries) {
		return
	}
	entry := s.menuEntries[s.menuIndex]
	s.soulX, s.soulY = entry.anchorX, entry.anchorY
}
func (s *demoScene) updateSoul(dt time.Duration) {
	if s.soul == nil {
		return
	}
	if !s.menuActive && s.battle != nil && s.battle.Phase != democomponents.BattlePhaseCommandSelect {
		s.soul.Update(dt)
		return
	}
	if !s.menuActive {
		previousButton := s.selectedButton
		if s.input.justPressed(ebiten.KeyArrowLeft) {
			s.selectedButton--
		}
		if s.input.justPressed(ebiten.KeyArrowRight) {
			s.selectedButton++
		}
		if s.selectedButton < 0 {
			s.selectedButton = len(s.buttonNames) - 1
		}
		if s.selectedButton >= len(s.buttonNames) {
			s.selectedButton = 0
		}
		if previousButton != s.selectedButton {
			s.playSFX(sndSqueak)
		}
		button := s.buttons[s.buttonNames[s.selectedButton]]
		if button != nil && button.Transform() != nil {
			s.soulX = button.Transform().X + 16
			s.soulY = 453
		}
		s.syncButtonSelection()
		if s.input.justPressed(ebiten.KeyZ) {
			s.menuConfirmed = false
			s.playSFX(sndSelect)
			s.openMenu()
		}
	} else {
		if s.input.justPressed(ebiten.KeyArrowUp) || s.input.justPressed(ebiten.KeyArrowLeft) {
			previousIndex := s.menuIndex
			s.menuIndex--
			if s.menuIndex < 0 {
				s.menuIndex = len(s.menuEntries) - 1
			}
			if s.menuIndex != previousIndex {
				s.playSFX(sndSqueak)
			}
			s.moveSoulToMenuEntry()
		}
		if s.input.justPressed(ebiten.KeyArrowDown) || s.input.justPressed(ebiten.KeyArrowRight) {
			previousIndex := s.menuIndex
			s.menuIndex++
			if s.menuIndex >= len(s.menuEntries) {
				s.menuIndex = 0
			}
			if s.menuIndex != previousIndex {
				s.playSFX(sndSqueak)
			}
			s.moveSoulToMenuEntry()
		}
		if s.input.justPressed(ebiten.KeyZ) {
			s.confirmMenuChoice()
		}
		if s.input.justPressed(ebiten.KeyX) {
			s.menuActive = false
			s.menuConfirmed = false
			s.playSFX(sndSqueak)
			s.closeMenu()
			s.syncButtonSelection()
		}
	}
	s.soul.Update(dt)
}

func (s *demoScene) confirmMenuChoice() {
	if s.menuIndex < 0 || s.menuIndex >= len(s.menuEntries) || s.battle == nil {
		return
	}
	selected := s.menuEntries[s.menuIndex].option
	if selected.prototype == "" {
		return
	}
	s.playSFX(sndSelect)
	s.selectedMenuOption = selected
	s.menuConfirmed = true
	s.menuActive = false
	s.closeMenu()
	s.battle.SetPhase(democomponents.BattlePhasePlayerAction)
	s.syncButtonSelection()
	s.runtime.Dialog.Clear()
	s.actionDialoguePending = true
	dialogue := s.selectedMenuOption.dialogue
	if len(dialogue) == 0 {
		dialogue = []string{"* " + s.selectedMenuOption.label + "."}
	}
	for _, line := range dialogue {
		s.runtime.Dialog.Show(line, battleDialogStyle(), sndText)
	}
	if !s.runtime.Dialog.Active() {
		s.finishPlayerActionDialogue()
	}
}

func (s *demoScene) finishPlayerActionDialogue() {
	if s.battle != nil {
		s.battle.Turn++
		s.battle.SetPhase(democomponents.BattlePhaseEnemyTurn)
	}
	s.actionDialoguePending = false
}

func (s *demoScene) playSFX(name string) {
	if player := s.runtime.Dialog.SoundPlayer; player != nil {
		_ = player.PlaySound(name, 0.5)
	}
}

func (s *demoScene) syncButtonSelection() {
	for i, name := range s.buttonNames {
		if button := s.buttons[name]; button != nil {
			democomponents.SetBattleButtonActive(button, s.buttonActive(i))
		}
	}
}

func (s *demoScene) buttonActive(index int) bool {
	if index != s.selectedButton || s.menuActive {
		return false
	}
	return s.battle == nil || s.battle.Phase == democomponents.BattlePhaseCommandSelect
}

func (s *demoScene) Draw(screen *ebiten.Image) {
	if s.frame == nil || s.frame.Bounds().Size() != screen.Bounds().Size() {
		s.frame = ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
	}
	s.frame.Fill(color.Black)
	s.runtime.Queues.ExecuteDraw(s.frame)
	if s.preview != nil {
		op := &ebiten.DrawImageOptions{}
		op.ColorScale.ScaleAlpha(0.3)
		s.frame.DrawImage(s.preview, op)
	}
	if s.soulVisible() {
		s.soul.Draw(s.frame, s.soulX-8, s.soulY-8, 1, 1, 0)
	}
	if s.phaseLabel != nil {
		s.phaseLabel.Draw(s.frame)
	}
	s.drawPlayerStatus(s.frame)
	s.camera.Apply(screen, func(target *ebiten.Image, ox, oy float64) {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(ox, oy)
		target.DrawImage(s.frame, op)
	})
}

func (s *demoScene) drawPlayerStatus(screen *ebiten.Image) {
	if s.playerCombatant == nil {
		return
	}
	s.playerNameText.Draw(screen)
	s.playerLevelText.Draw(screen)
	s.playerHPText.Draw(screen)

	hp := s.playerCombatant.HP
	maxHP := s.playerCombatant.MaxHP
	value := fmt.Sprintf("%d/%d", hp, maxHP)
	if value != s.playerHPValue {
		s.playerHPValue = value
		s.playerHPValueText.SetText(value)
	}
	barWidth := hpBarWidth(maxHP)
	valueX := s.playerHPText.X + float64(barWidth) + 52
	if valueX != s.playerHPValueX {
		s.playerHPValueX = valueX
		s.playerHPValueText.SetPosition(valueX, 384)
	}
	s.playerHPValueText.Draw(screen)
	drawPlayerHPBar(screen, float32(s.playerHPText.X+40), 403, barWidth, 17, hp, maxHP)
}

func statusTextWidth(font *render.AnimatedIcon, value string, scale float64) float64 {
	fallback := 0.0
	if frame := font.CurrentState.CurrentFrameRef; frame != nil {
		fallback = float64(frame.Image.Bounds().Dx())
	}
	width := 0.0
	for _, character := range value {
		if err := font.SetIconState(string(character)); err != nil {
			width += fallback
			continue
		}
		frame := font.CurrentState.CurrentFrameRef
		if frame == nil {
			width += fallback
			continue
		}
		advance := frame.Advance
		if advance <= 0 {
			advance = float64(frame.Image.Bounds().Dx())
		}
		width += advance
	}
	return width * scale
}

func hpBarWidth(maxHP int) float32 {
	if maxHP <= 0 {
		return 0
	}
	width := float32(maxHP) * 1.5
	return width
}

func drawPlayerHPBar(screen *ebiten.Image, x, y, width, height float32, hp, maxHP int) {
	if width <= 0 || height <= 0 {
		return
	}
	vector.DrawFilledRect(screen, x, y, width, height, color.RGBA{R: 255, A: 255}, false)
	if hp < 0 {
		hp = 0
	}
	if hp > maxHP {
		hp = maxHP
	}
	if maxHP <= 0 {
		return
	}
	filledWidth := width * float32(hp) / float32(maxHP)
	if filledWidth > 0 {
		vector.DrawFilledRect(screen, x, y, filledWidth, height, color.RGBA{R: 255, G: 255, A: 255}, false)
	}
}
