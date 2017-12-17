package parser

import (
	"gopheros/device/acpi/aml/entity"
	"gopheros/kernel/kfmt"
	"io"
)

// resolveSymbolRefs is invoked by the parser after the AML stream has been
// parsed to resolve any forward declarations that an entity may require.
func resolveSymbolRefs(errWriter io.Writer, e entity.Entity, rootNs entity.Container) bool {
	var ok bool

	switch ent := e.(type) {
	case *entity.Field:
		if ent.Region, ok = scopeFind(ent.Parent(), rootNs, ent.RegionName).(*entity.Region); !ok {
			kfmt.Fprintf(errWriter, "could not resolve referenced field region: %s\n", ent.RegionName)
			return false
		}
	case *entity.IndexField:
		if ent.IndexReg, ok = scopeFind(ent.Parent(), rootNs, ent.IndexRegName).(*entity.FieldUnit); !ok {
			kfmt.Fprintf(errWriter, "could not resolve referenced index register: %s\n", ent.IndexRegName)
			return false
		}

		if ent.DataReg, ok = scopeFind(ent.Parent(), rootNs, ent.DataRegName).(*entity.FieldUnit); !ok {
			kfmt.Fprintf(errWriter, "could not resolve referenced data register: %s\n", ent.DataRegName)
			return false
		}
	case *entity.BankField:
		if ent.Region, ok = scopeFind(ent.Parent(), rootNs, ent.RegionName).(*entity.Region); !ok {
			kfmt.Fprintf(errWriter, "could not resolve referenced field region: %s\n", ent.RegionName)
			return false
		}

		if ent.BankFieldUnit, ok = scopeFind(ent.Parent(), rootNs, ent.BankFieldUnitName).(*entity.FieldUnit); !ok {
			kfmt.Fprintf(errWriter, "could not resolve referenced bank register field: %s\n", ent.BankFieldUnitName)
			return false
		}
	case *entity.FieldUnit:
		if ent.ConnectionName != "" {
			if ent.Connection = scopeFind(ent.Parent(), rootNs, ent.ConnectionName); ent.Connection == nil {
				kfmt.Fprintf(errWriter, "[field unit %s] could not resolve connection reference: %s\n", ent.Name(), ent.ConnectionName)
				return false
			}
		}
	case *entity.Reference:
		if ent.Target = scopeFind(ent.Parent(), rootNs, ent.TargetName); ent.Target == nil {
			kfmt.Fprintf(errWriter, "could not resolve referenced symbol: %s (parent: %s)\n", ent.TargetName, ent.Parent().Name())
			return false
		}
	case *entity.Invocation:
		if ent.MethodDef, ok = scopeFind(ent.Parent(), rootNs, ent.MethodName).(*entity.Method); !ok {
			kfmt.Fprintf(errWriter, "could not resolve method invocation to: %s (parent: %s)\n", ent.MethodName, ent.Parent().Name())
			return false
		}
	}
	return true
}
