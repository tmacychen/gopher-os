package entity

import (
	"reflect"
	"testing"
)

func TestEntityMethods(t *testing.T) {
	namedConst := NewConst(OpDwordPrefix, 42, "foo")
	namedConst.SetName("TAG0")

	specs := []struct {
		ent     Entity
		expOp   AMLOpcode
		expName string
	}{
		{NewGeneric(OpNoop, 42), OpNoop, ""},
		{NewGenericNamed(OpAcquire, 42), OpAcquire, ""},
		{namedConst, OpDwordPrefix, "TAG0"},
		{NewScope(OpScope, 42, "_SB_"), OpScope, "_SB_"},
		{NewBuffer(42), OpBuffer, ""},
		{NewBufferField(OpCreateByteField, 42, 8), OpCreateByteField, ""},
		{NewField(42), OpField, ""},
		{NewIndexField(42), OpIndexField, ""},
		{NewBankField(42), OpBankField, ""},
		{NewReference(42, "TRG0"), OpName, ""},
		{NewMethod(42, "FOO0"), OpMethod, "FOO0"},
		{NewInvocation(42, "MTH0"), OpMethodInvocation, ""},
		{NewMutex(42), OpMutex, ""},
		{NewDevice(42, "DEV0"), OpDevice, "DEV0"},
		{NewProcessor(42, "CPU0"), OpProcessor, "CPU0"},
		{NewPowerResource(42, "POW0"), OpPowerRes, "POW0"},
		{NewThermalZone(42, "THE0"), OpThermalZone, "THE0"},
		{NewEvent(42), OpEvent, ""},
		{NewRegion(42), OpOpRegion, ""},
		{NewFieldUnit(42, "FOO0"), OpFieldUnit, "FOO0"},
	}

	t.Run("opcode and name getter", func(t *testing.T) {
		for specIndex, spec := range specs {
			if got := spec.ent.Opcode(); got != spec.expOp {
				t.Errorf("[spec %d] expected to get back opcode %d; got %d", specIndex, spec.expOp, got)
			}

			if got := spec.ent.Name(); got != spec.expName {
				t.Errorf("[spec %d] expected to get name: %q; got %q", specIndex, spec.expName, got)
			}
		}
	})

	t.Run("table handle getter", func(t *testing.T) {
		exp := uint8(42)
		for specIndex, spec := range specs {
			if got := spec.ent.TableHandle(); got != exp {
				t.Errorf("[spec %d] expected to get back handle %d; got %d", specIndex, exp, got)
			}
		}
	})

	t.Run("append/remove/get parent methods", func(t *testing.T) {
		parent := NewScope(OpScope, 2, "_SB_")
		parent.name = `\`

		for specIndex, spec := range specs {
			parent.Append(spec.ent)
			if got := spec.ent.Parent(); got != parent {
				t.Errorf("[spec %d] expected to get back parent %v; got %v", specIndex, parent, got)
			}

			if got := parent.Last(); got != spec.ent {
				t.Errorf("[spec %d] expected parent's last entity to be the one just appended", specIndex)
			}

			parent.Remove(spec.ent)
		}

		if got := len(parent.Children()); got != 0 {
			t.Fatalf("expected parent not to have any child nodes; got %d", got)
		}
	})
}

func TestEntityArgAssignment(t *testing.T) {
	specs := []struct {
		ent         Entity
		argList     []interface{}
		expArgList  []interface{}
		limitedArgs bool
	}{
		{
			NewGeneric(1, 2),
			[]interface{}{"foo", 1, "bar"},
			[]interface{}{"foo", 1, "bar"},
			false,
		},
		{
			NewGenericNamed(1, 2),
			[]interface{}{"foo", 1, "bar"},
			[]interface{}{1, "bar"}, // GenericNamed uses arg0 as the name
			false,
		},
		{
			NewConst(1, 2, 3),
			[]interface{}{"foo"},
			nil, // Const populates its internal state using the arg 0
			true,
		},
		{
			NewBuffer(2),
			[]interface{}{1, []byte{}},
			nil, // Buffer populates its internal state using the first 2 args
			true,
		},
		{
			NewBufferField(OpCreateDWordField, 2, 32),
			[]interface{}{"a", "b", "c"},
			nil, // Buffer populates its internal state using the first 3 args (opCreateDwordField)
			false,
		},
		{
			NewBufferField(1, 2, 0),
			[]interface{}{"a", "b", 10, "c"},
			nil, // Buffer populates its internal state using the first 4 args (opCreateField)
			true,
		},
		{
			NewRegion(2),
			[]interface{}{"REG0", uint64(0x4), 0, 10},
			nil, // Region populates its internal state using the first 4 args
			true,
		},
		{
			NewMutex(2),
			[]interface{}{"MUT0", uint64(1)},
			nil, // Mutex populates its internal state using the first 2 args
			true,
		},
		{
			NewProcessor(2, "CPU0"),
			[]interface{}{uint64(1), uint64(0xdeadc0de), uint64(0)},
			nil, // Processor populates its internal state using the first 3 args
			true,
		},
		{
			NewPowerResource(2, "POW0"),
			[]interface{}{uint64(2), uint64(1)},
			nil, // PowerResource populates its internal state using the first 2 args
			true,
		},
		{
			NewMethod(2, "MTH0"),
			[]interface{}{"arg0 ignored", uint64(0x42)},
			nil, // Method populates its internal state using the first 2 args
			true,
		},
	}

nextSpec:
	for specIndex, spec := range specs {
		for i, arg := range spec.argList {
			if !spec.ent.SetArg(uint8(i), arg) {
				t.Errorf("[spec %d] error setting arg %d", specIndex, i)
				continue nextSpec
			}
		}

		if spec.limitedArgs {
			if spec.ent.SetArg(uint8(len(spec.argList)), nil) {
				t.Errorf("[spec %d] expected additional calls to setArg to return false", specIndex)
				continue nextSpec
			}
		}

		if got := spec.ent.Args(); !reflect.DeepEqual(got, spec.expArgList) {
			t.Errorf("[spec %d] expected to get back arg list %v; got %v", specIndex, spec.expArgList, got)
		}
	}
}
