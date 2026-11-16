package components

import (
	"os"
	"testing"

	enginecomponents "github.com/mcbalaam/ebitter/components"
	"gopkg.in/yaml.v3"
)

func TestBattleComponentsRegistered(t *testing.T) {
	for _, name := range []string{
		"combatant", "enemyController", "battleAction", "item", "inventory",
		"attackPattern", "projectile", "soulController", "battleController",
	} {
		if !enginecomponents.Registered(name) {
			t.Errorf("component %q is not registered", name)
		}
	}
}

func TestCombatantDamageAndHealing(t *testing.T) {
	c := &Combatant{HP: 10, MaxHP: 10, Defense: 3}
	if got := c.Damage(7); got != 4 || c.HP != 6 {
		t.Fatalf("Damage(7) = %d, HP = %d; want 4, 6", got, c.HP)
	}
	if got := c.Heal(20); got != 4 || c.HP != 10 {
		t.Fatalf("Heal(20) = %d, HP = %d; want 4, 10", got, c.HP)
	}
	if got := c.Damage(100); got != 10 || !c.Defeated() {
		t.Fatalf("lethal damage = %d, defeated = %v; want 10, true", got, c.Defeated())
	}
}

func TestInventoryStacksAndRemovesItems(t *testing.T) {
	i := &Inventory{}
	i.Add("pie", 2)
	i.Add("pie", 3)
	if got := i.Count("pie"); got != 5 {
		t.Fatalf("Count(pie) = %d, want 5", got)
	}
	if !i.Remove("pie", 5) || i.Count("pie") != 0 {
		t.Fatal("expected to remove the full pie stack")
	}
	if i.Remove("pie", 1) {
		t.Fatal("removing a missing item should fail")
	}
}

func TestBattleComponentConfigValidation(t *testing.T) {
	valid := &yaml.Node{}
	if err := yaml.Unmarshal([]byte("name: Check\ntarget: enemy\neffects:\n  - type: damage\n    amount: 4\n"), valid); err != nil {
		t.Fatal(err)
	}
	if err := enginecomponents.ValidateComponent("battleAction", valid.Content[0]); err != nil {
		t.Fatalf("valid action rejected: %v", err)
	}

	invalid := &yaml.Node{}
	if err := yaml.Unmarshal([]byte("durationMS: 100\nevents:\n  - atMS: 80\n    projectile: spark\n  - atMS: 20\n    projectile: spark\n"), invalid); err != nil {
		t.Fatal(err)
	}
	if err := enginecomponents.ValidateComponent("attackPattern", invalid.Content[0]); err == nil {
		t.Fatal("expected unsorted attack events to be rejected")
	}
}

func TestBattleControllerDefaultsPhase(t *testing.T) {
	c := &BattleController{}
	c.Init(nil)
	if c.Phase != BattlePhaseCommandSelect {
		t.Fatalf("initial phase = %q, want %q", c.Phase, BattlePhaseCommandSelect)
	}
}

func TestTrainingDummyPrototypeExamplesLoad(t *testing.T) {
	library := enginecomponents.NewBlueprintLibrary(os.DirFS("../prototypes"), ".")
	for _, id := range []string{
		"trainingDummy",
		"trainingDummy_check",
		"trainingDummy_slowShot",
		"trainingDummy_spark",
	} {
		if _, err := library.Load(id); err != nil {
			t.Errorf("load prototype %q: %v", id, err)
		}
	}
}
