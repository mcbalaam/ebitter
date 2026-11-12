package components

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type registryTestComponent string

func (c registryTestComponent) Type() string { return string(c) }

func TestRegisterRejectsInvalidAndDuplicateRegistrations(t *testing.T) {
	factory := func(*yaml.Node) (Component, error) {
		return registryTestComponent("test_registry_component"), nil
	}

	validator := validateKeys()
	if err := Register("  ", 0, factory, validator); err == nil {
		t.Fatal("Register accepted an empty name")
	}
	if err := Register("test_registry_nil_factory", 0, nil, validator); err == nil {
		t.Fatal("Register accepted a nil factory")
	}
	if err := Register("test_registry_component", 0, factory, validator); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := Register("test_registry_component", 1, factory, validator); err == nil {
		t.Fatal("Register accepted a duplicate name")
	}
}

func TestNewComponentChecksFactoryResult(t *testing.T) {
	if err := Register("test_registry_wrong_type", 0, func(*yaml.Node) (Component, error) {
		return registryTestComponent("not_test_registry_wrong_type"), nil
	}, validateKeys()); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if _, err := NewComponent("test_registry_wrong_type", nil); err == nil || !strings.Contains(err.Error(), "returned component type") {
		t.Fatalf("NewComponent error = %v, want type mismatch", err)
	}

	if err := Register("test_registry_nil_result", 0, func(*yaml.Node) (Component, error) {
		return nil, nil
	}, validateKeys()); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if _, err := NewComponent("test_registry_nil_result", nil); err == nil || !strings.Contains(err.Error(), "returned nil") {
		t.Fatalf("NewComponent error = %v, want nil result", err)
	}
}
