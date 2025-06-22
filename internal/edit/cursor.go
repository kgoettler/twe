// Package edit provides cursor and position types for navigating a table-like data structure.
package edit

// position represents a row and column index within a table.
type position struct {
	row int // Current row index
	col int // Current column index
}

// Up moves the position up by one row, if not already at the top.
func (p *position) Up() {
	if p.row > 0 {
		p.row--
	}
}

// Down moves the position down by one row.
func (p *position) Down() {
	p.row++
}

// Left moves the position left by one column, if not already at the leftmost column.
func (p *position) Left() {
	if p.col > 0 {
		p.col--
	}
}

// Right moves the position right by one column.
func (p *position) Right() {
	p.col++
}

// cursor tracks the current position and table dimensions for navigation.
type cursor struct {
	pos   *position // Current position in the table
	nrows int       // Number of rows in the table
	ncols int       // Number of columns in the table
}

// NewCursor creates a new cursor for a table with the given number of rows and columns.
func NewCursor(nrows int, ncols int) cursor {
	return cursor{pos: &position{}, nrows: nrows, ncols: ncols}
}

// AddRow increments the number of rows tracked by the cursor.
func (c *cursor) AddRow() {
	c.nrows++
}

// RemoveRow decrements the number of rows and adjusts the cursor position if needed.
func (c *cursor) RemoveRow() {
	if c.nrows > 0 {
		c.nrows--
	}
	if c.pos.row >= c.nrows {
		c.pos.row--
	}
}

// AddCol increments the number of columns tracked by the cursor.
func (c *cursor) AddCol() {
	c.ncols++
}

// RemoveCol decrements the number of columns tracked by the cursor.
func (c *cursor) RemoveCol() {
	if c.ncols > 0 {
		c.ncols--
	}
}

// GetRow returns the current row index of the cursor.
func (c *cursor) GetRow() int {
	return c.pos.row
}

// GetCol returns the current column index of the cursor.
func (c *cursor) GetCol() int {
	return c.pos.col
}

// Up moves the cursor up by one row, if possible.
func (c *cursor) Up() {
	c.pos.Up()
}

// Down moves the cursor down by one row, if not at the last row.
func (c *cursor) Down() {
	if c.pos.row < c.nrows-1 {
		c.pos.Down()
	}
}

// Left moves the cursor left by one column, if possible.
func (c *cursor) Left() {
	c.pos.Left()
}

// Right moves the cursor right by one column, if not at the last column.
func (c *cursor) Right() {
	if c.pos.col < c.ncols-1 {
		c.pos.Right()
	}
}
