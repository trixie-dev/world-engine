package component_metadata_test

import (
	"testing"

	"gotest.tools/v3/assert"
	"pkg.world.dev/world-engine/cardinal/ecs/entity"
	"pkg.world.dev/world-engine/cardinal/ecs/filter"

	"pkg.world.dev/world-engine/cardinal/ecs"
	"pkg.world.dev/world-engine/cardinal/ecs/archetype"
	"pkg.world.dev/world-engine/cardinal/ecs/component"
	"pkg.world.dev/world-engine/cardinal/ecs/component_metadata"

	"pkg.world.dev/world-engine/cardinal/ecs/storage"
)

type ComponentDataA struct {
	Value string
}

func (ComponentDataA) Name() string { return "a" }

type ComponentDataB struct {
	Value string
}

func (ComponentDataB) Name() string { return "b" }

func TestComponents(t *testing.T) {
	world := ecs.NewTestWorld(t)
	ecs.MustRegisterComponent[ComponentDataA](world)
	ecs.MustRegisterComponent[ComponentDataB](world)

	ca, err := world.GetComponentByName("a")
	assert.NilError(t, err)
	cb, err := world.GetComponentByName("b")
	assert.NilError(t, err)

	tests := []*struct {
		comps    []component_metadata.IComponentMetaData
		archID   archetype.ID
		entityID entity.ID
		Value    string
	}{
		{
			[]component_metadata.IComponentMetaData{ca},
			0,
			0,
			"a",
		},
		{
			[]component_metadata.IComponentMetaData{ca, cb},
			1,
			0,
			"b",
		},
	}

	storeManager := world.StoreManager()
	for _, tt := range tests {
		entityID, err := storeManager.CreateEntity(tt.comps...)
		assert.NilError(t, err)
		tt.entityID = entityID
		tt.archID, err = storeManager.GetArchIDForComponents(tt.comps)
		assert.NilError(t, err)
	}

	for _, tt := range tests {
		componentsForArchID := storeManager.GetComponentTypesForArchID(tt.archID)
		for _, comp := range tt.comps {
			ok := filter.MatchComponentMetaData(componentsForArchID, comp)
			if !ok {
				t.Errorf("the archtype ID %d shoudl contain the component %dd", tt.archID, comp.ID())
			}
			iface, err := storeManager.GetComponentForEntity(comp, tt.entityID)
			assert.NilError(t, err)

			switch component := iface.(type) {
			case ComponentDataA:
				component.Value = tt.Value
				assert.NilError(t, storeManager.SetComponentForEntity(ca, tt.entityID, component))
			case ComponentDataB:
				component.Value = tt.Value
				assert.NilError(t, storeManager.SetComponentForEntity(cb, tt.entityID, component))
			default:
				assert.Check(t, false, "unknown component type: %v", iface)
			}
		}
	}

	target := tests[0]

	srcArchIdx := target.archID
	var dstArchIdx archetype.ID = 1

	assert.NilError(t, storeManager.AddComponentToEntity(cb, target.entityID))

	gotComponents, err := storeManager.GetComponentTypesForEntity(target.entityID)
	assert.NilError(t, err)
	gotArchID, err := storeManager.GetArchIDForComponents(gotComponents)
	assert.NilError(t, err)
	assert.Check(t, gotArchID != srcArchIdx, "the archetype ID should be different after adding a component")

	gotIDs, err := storeManager.GetEntitiesForArchID(srcArchIdx)
	assert.NilError(t, err)
	assert.Equal(t, 0, len(gotIDs), "there should be no entities in the archetype ID %d", srcArchIdx)

	gotIDs, err = storeManager.GetEntitiesForArchID(dstArchIdx)
	assert.NilError(t, err)
	assert.Equal(t, 2, len(gotIDs), "there should be 2 entities in the archetype ID %d", dstArchIdx)

	iface, err := storeManager.GetComponentForEntity(ca, target.entityID)
	assert.NilError(t, err)

	got, ok := iface.(ComponentDataA)
	assert.Check(t, ok, "component %v is of wrong type", iface)
	assert.Equal(t, got.Value, target.Value, "component should have value of %q got %q", target.Value, got.Value)
}

type foundComp struct{}
type notFoundComp struct{}

func (_ foundComp) Name() string {
	return "foundComp"
}

func (_ notFoundComp) Name() string {
	return "notFoundComp"
}

func TestErrorWhenAccessingComponentNotOnEntity(t *testing.T) {
	world := ecs.NewTestWorld(t)
	ecs.MustRegisterComponent[foundComp](world)
	ecs.MustRegisterComponent[notFoundComp](world)

	wCtx := ecs.NewWorldContext(world)
	id, err := component.Create(wCtx, foundComp{})
	assert.NilError(t, err)
	_, err = component.GetComponent[notFoundComp](wCtx, id)
	assert.ErrorIs(t, err, storage.ErrorComponentNotOnEntity)
}

type ValueComponent struct {
	Val int
}

func (ValueComponent) Name() string {
	return "ValueComponent"
}

func TestMultipleCallsToCreateSupported(t *testing.T) {

	world := ecs.NewTestWorld(t)
	assert.NilError(t, ecs.RegisterComponent[ValueComponent](world))

	wCtx := ecs.NewWorldContext(world)
	id, err := component.Create(wCtx, ValueComponent{})
	assert.NilError(t, err)

	assert.NilError(t, component.SetComponent[ValueComponent](wCtx, id, &ValueComponent{99}))

	val, err := component.GetComponent[ValueComponent](wCtx, id)
	assert.NilError(t, err)
	assert.Equal(t, 99, val.Val)

	_, err = component.Create(wCtx, ValueComponent{})

	val, err = component.GetComponent[ValueComponent](wCtx, id)
	assert.NilError(t, err)
	assert.Equal(t, 99, val.Val)
}
