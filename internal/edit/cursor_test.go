package edit_test

import (
	"testing"

	. "github.com/kgoettler/twe/internal/edit"
)

func TestCursorNavigation(t *testing.T) {
	c := NewCursor(3, 3)
	if c.GetRow() != 0 || c.GetCol() != 0 {
		t.Errorf("Initial cursor position should be (0,0)")
	}
	c.Down()
	c.Down()
	c.Down() // should not go past nrows-1
	if c.GetRow() != 2 {
		t.Errorf("Down() boundary failed: expected row 2, got %d", c.GetRow())
	}
	c.Up()
	if c.GetRow() != 1 {
		t.Errorf("Up() failed: expected row 1, got %d", c.GetRow())
	}
	c.Right()
	c.Right()
	c.Right() // should not go past ncols-1
	if c.GetCol() != 2 {
		t.Errorf("Right() boundary failed: expected col 2, got %d", c.GetCol())
	}
	c.Left()
	if c.GetCol() != 1 {
		t.Errorf("Left() failed: expected col 1, got %d", c.GetCol())
	}
}

func TestCursorAddRemoveRowCol(t *testing.T) {
	c := NewCursor(2, 2)
	if c.GetRow() != 0 || c.GetCol() != 0 {
		t.Errorf("Initial cursor position should be (0,0)")
	}

	c.AddRow()
	c.Down()
	c.Down()
	if r := c.GetRow(); r != 2 {
		t.Errorf("After AddRow and Down, expected row 2, got %d", r)
	}

	c.RemoveRow()
	if r := c.GetRow(); r != 1 {
		t.Errorf("After RemoveRow, expected row 1, got %d", r)
	}

	c.AddCol()
	c.Right()
	c.Right()
	if c := c.GetCol(); c != 2 {
		t.Errorf("After AddCol and Right, expected col 2, got %d", c)
	}

	c.RemoveCol()
	if c := c.GetCol(); c != 1 {
		t.Errorf("After RemoveCol, expected col 1, got %d", c)
	}
}
