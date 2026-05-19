package agl

import "testing"

func TestOperandString(t *testing.T) {
	if got := (Operand{IsField: true, Field: "task.uid"}).String(); got != "task.uid" {
		t.Errorf("field operand: got %q", got)
	}
	if got := (Operand{Value: Value{Type: ValueInt, Int: 7}}).String(); got != "7" {
		t.Errorf("literal operand: got %q", got)
	}
}
