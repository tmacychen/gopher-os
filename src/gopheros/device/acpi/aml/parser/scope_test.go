package parser

import (
	"gopheros/device/acpi/aml/entity"
	"reflect"
	"testing"
)

func TestScopeResolvePath(t *testing.T) {
	scopeMap := genTestScopes()

	specs := []struct {
		curScope   entity.Container
		pathExpr   string
		wantParent entity.Entity
		wantName   string
	}{
		{
			scopeMap["IDE0"].(entity.Container),
			`\_SB_`,
			scopeMap[`\`],
			"_SB_",
		},
		{
			scopeMap["IDE0"].(entity.Container),
			`^FOO`,
			scopeMap[`PCI0`],
			"FOO",
		},
		{
			scopeMap["IDE0"].(entity.Container),
			`^^FOO`,
			scopeMap[`_SB_`],
			"FOO",
		},
		{
			scopeMap["IDE0"].(entity.Container),
			`_ADR`,
			scopeMap[`IDE0`],
			"_ADR",
		},
		// Paths with dots
		{
			scopeMap["IDE0"].(entity.Container),
			`\_SB_.PCI0.IDE0._ADR`,
			scopeMap[`IDE0`],
			"_ADR",
		},
		{
			scopeMap["PCI0"].(entity.Container),
			`IDE0._ADR`,
			scopeMap[`IDE0`],
			"_ADR",
		},
		{
			scopeMap["PCI0"].(entity.Container),
			`_CRS`,
			scopeMap[`PCI0`],
			"_CRS",
		},
		// Bad queries
		{
			scopeMap["PCI0"].(entity.Container),
			`FOO.BAR.BAZ`,
			nil,
			"",
		},
		{
			scopeMap["PCI0"].(entity.Container),
			``,
			nil,
			"",
		},
		{
			scopeMap["PCI0"].(entity.Container),
			`\`,
			nil,
			"",
		},
		{
			scopeMap["PCI0"].(entity.Container),
			`^^^^^^^^^BADPATH`,
			nil,
			"",
		},
	}

	root := scopeMap[`\`].(entity.Container)
	for specIndex, spec := range specs {
		gotParent, gotName := scopeResolvePath(spec.curScope, root, spec.pathExpr)
		if !reflect.DeepEqual(gotParent, spec.wantParent) {
			t.Errorf("[spec %d] expected lookup to return %#v; got %#v", specIndex, spec.wantParent, gotParent)
			continue
		}

		if gotName != spec.wantName {
			t.Errorf("[spec %d] expected lookup to return node name %q; got %q", specIndex, spec.wantName, gotName)
		}
	}
}

func TestScopeFind(t *testing.T) {
	scopeMap := genTestScopes()

	specs := []struct {
		curScope entity.Container
		lookup   string
		want     entity.Entity
	}{
		// Search rules do not apply for these cases
		{
			scopeMap["PCI0"].(entity.Container),
			`\`,
			scopeMap[`\`],
		},
		{
			scopeMap["PCI0"].(entity.Container),
			"IDE0._ADR",
			scopeMap["_ADR"],
		},
		{
			scopeMap["IDE0"].(entity.Container),
			"^^PCI0.IDE0._ADR",
			scopeMap["_ADR"],
		},
		{
			scopeMap["IDE0"].(entity.Container),
			`\_SB_.PCI0.IDE0._ADR`,
			scopeMap["_ADR"],
		},
		{
			scopeMap["IDE0"].(entity.Container),
			`\_SB_.PCI0`,
			scopeMap["PCI0"],
		},
		{
			scopeMap["IDE0"].(entity.Container),
			`^`,
			scopeMap["PCI0"],
		},
		// Bad queries
		{
			scopeMap["_SB_"].(entity.Container),
			"PCI0.USB._CRS",
			nil,
		},
		{
			scopeMap["IDE0"].(entity.Container),
			"^^^^^^^^^^^^^^^^^^^",
			nil,
		},
		{
			scopeMap["IDE0"].(entity.Container),
			`^^^^^^^^^^^FOO`,
			nil,
		},
		{
			scopeMap["IDE0"].(entity.Container),
			"FOO",
			nil,
		},
		{
			scopeMap["IDE0"].(entity.Container),
			"",
			nil,
		},
		// Search rules apply for these cases
		{
			scopeMap["IDE0"].(entity.Container),
			"_CRS",
			scopeMap["_CRS"],
		},
	}

	root := scopeMap[`\`].(entity.Container)
	for specIndex, spec := range specs {
		if got := scopeFind(spec.curScope, root, spec.lookup); !reflect.DeepEqual(got, spec.want) {
			t.Errorf("[spec %d] expected lookup to return %#v; got %#v", specIndex, spec.want, got)
		}
	}
}

func genTestScopes() map[string]entity.Entity {
	// Setup the example tree from page 252 of the acpi 6.2 spec
	// \
	//  SB
	//    \
	//     PCI0
	//         | _CRS
	//         \
	//          IDE0
	//              | _ADR
	ideScope := entity.NewScope(entity.OpScope, 42, `IDE0`)
	pciScope := entity.NewScope(entity.OpScope, 42, `PCI0`)
	sbScope := entity.NewScope(entity.OpScope, 42, `_SB_`)
	rootScope := entity.NewScope(entity.OpScope, 42, `\`)

	adr := entity.NewMethod(42, `_ADR`)
	crs := entity.NewMethod(42, `_CRS`)

	// Setup tree
	ideScope.Append(adr)
	pciScope.Append(crs)
	pciScope.Append(ideScope)
	sbScope.Append(pciScope)
	rootScope.Append(sbScope)

	return map[string]entity.Entity{
		"IDE0": ideScope,
		"PCI0": pciScope,
		"_SB_": sbScope,
		"\\":   rootScope,
		"_ADR": adr,
		"_CRS": crs,
	}
}
