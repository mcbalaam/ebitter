package components

import (
	"bytes"
	"fmt"

	enginecomponents "github.com/mcbalaam/ebitter/components"
	"gopkg.in/yaml.v3"
)

func registerBattleComponent[T enginecomponents.Component](name string, order int, newComponent func() T, validate func(T) error) {
	err := enginecomponents.Register(name, order, func(node *yaml.Node) (enginecomponents.Component, error) {
		component := newComponent()
		if node != nil {
			if err := node.Decode(component); err != nil {
				return nil, err
			}
		}
		return component, nil
	}, func(node *yaml.Node) error {
		if node == nil {
			return nil
		}
		if node.Kind != yaml.MappingNode {
			return fmt.Errorf("expected mapping")
		}
		copyNode := *node
		copyNode.Content = make([]*yaml.Node, 0, len(node.Content))
		for n := 0; n+1 < len(node.Content); n += 2 {
			if node.Content[n].Value != "order" {
				copyNode.Content = append(copyNode.Content, node.Content[n], node.Content[n+1])
			}
		}
		data, err := yaml.Marshal(&copyNode)
		if err != nil {
			return err
		}
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(true)
		component := newComponent()
		if err := decoder.Decode(component); err != nil {
			return err
		}
		return validate(component)
	})
	if err != nil {
		panic(err)
	}
}
