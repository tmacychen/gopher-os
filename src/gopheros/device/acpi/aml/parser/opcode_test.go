package parser

import (
	"gopheros/device/acpi/aml/entity"
	"testing"
)

func TestOpcodeToString(t *testing.T) {
	if exp, got := "Acquire", opName(entity.OpAcquire); got != exp {
		t.Fatalf("expected opName(OpAcquire) to return %q; got %q", exp, got)
	}

	if exp, got := "unknown", opName(entity.AMLOpcode(0xffff)); got != exp {
		t.Fatalf("expected opName(invalid) to return %q; got %q", exp, got)
	}
}

// TestFindUnmappedOpcodes is a helper test that pinpoints opcodes that have
// not yet been mapped via an opcode table.
func TestFindUnmappedOpcodes(t *testing.T) {
	for opIndex, opRef := range opcodeMap {
		if opRef != badOpcode {
			continue
		}

		for tabIndex, info := range opcodeTable {
			if uint16(info.op) == uint16(opIndex) {
				t.Errorf("set opcodeMap[0x%02x] = 0x%02x // %s\n", opIndex, tabIndex, opName(info.op))
				break
			}
		}
	}

	for opIndex, opRef := range extendedOpcodeMap {
		// 0xff (opOnes) is defined in opcodeTable
		if opRef != badOpcode || opIndex == 0 {
			continue
		}

		opIndex += 0xff
		for tabIndex, info := range opcodeTable {
			if uint16(info.op) == uint16(opIndex) {
				t.Errorf("set extendedOpcodeMap[0x%02x] = 0x%02x // %s\n", opIndex-0xff, tabIndex, opName(info.op))
				break
			}
		}
	}
}
