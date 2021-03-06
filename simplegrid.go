// 25 february 2014

package ui

import (
	"fmt"
)

// A SimpleGrid arranges Controls in a two-dimensional grid.
// The height of each row and the width of each column is the maximum preferred height and width (respectively) of all the controls in that row or column (respectively).
// Controls are aligned to the top left corner of each cell.
// All Controls in a SimpleGrid maintain their preferred sizes by default; if a Control is marked as being "filling", it will be sized to fill its cell.
// Even if a Control is marked as filling, its preferred size is used to calculate cell sizes.
// One Control can be marked as "stretchy": when the Window containing the SimpleGrid is resized, the cell containing that Control resizes to take any remaining space; its row and column are adjusted accordingly (so other filling controls in the same row and column will fill to the new height and width, respectively).
// A stretchy Control implicitly fills its cell.
// All cooridnates in a SimpleGrid are given in (row,column) form with (0,0) being the top-left cell.
//
// As a special rule, to ensure proper appearance, non-standalone Labels are automatically made filling.
type SimpleGrid interface {
	Control

	// SetFilling marks the given Control of the SimpleGrid as filling its cell instead of staying at its preferred size.
	// It panics if the given coordinate is invalid.
	SetFilling(row int, column int)

	// SetStretchy marks the given Control of the SimpleGrid as stretchy.
	// Stretchy implies filling.
	// Only one control can be stretchy per SimpleGrid; calling SetStretchy multiple times merely changes which control is stretchy (preserving the previous filling value).
	// It panics if the given coordinate is invalid.
	SetStretchy(row int, column int)
}

type simpleGrid struct {
	controls                 [][]Control
	filling                  [][]bool
	stretchyrow, stretchycol int
	stretchyfill             bool
	widths, heights          [][]int // caches to avoid reallocating each time
	rowheights, colwidths    []int
}

// NewSimpleGrid creates a new SimpleGrid with the given Controls.
// NewSimpleGrid needs to know the number of Controls in a row (alternatively, the number of columns); it will determine the number in a column from the number of Controls given.
// NewSimpleGrid panics if not given a full grid of Controls.
// Example:
// 	grid := NewSimpleGrid(3,
// 		control00, control01, control02,
// 		control10, control11, control12,
// 		control20, control21, control22)
func NewSimpleGrid(nPerRow int, controls ...Control) SimpleGrid {
	if len(controls)%nPerRow != 0 {
		panic(fmt.Errorf("incomplete simpleGrid given to NewSimpleGrid() (not enough controls to evenly divide %d controls into rows of %d controls each)", len(controls), nPerRow))
	}
	nRows := len(controls) / nPerRow
	cc := make([][]Control, nRows)
	cf := make([][]bool, nRows)
	cw := make([][]int, nRows)
	ch := make([][]int, nRows)
	i := 0
	for row := 0; row < nRows; row++ {
		cc[row] = make([]Control, nPerRow)
		cf[row] = make([]bool, nPerRow)
		cw[row] = make([]int, nPerRow)
		ch[row] = make([]int, nPerRow)
		for x := 0; x < nPerRow; x++ {
			cc[row][x] = controls[i]
			if l, ok := controls[i].(Label); ok && !l.isStandalone() {
				cf[row][x] = true
			}
			i++
		}
	}
	return &simpleGrid{
		controls:    cc,
		filling:     cf,
		stretchyrow: -1,
		stretchycol: -1,
		widths:      cw,
		heights:     ch,
		rowheights:  make([]int, nRows),
		colwidths:   make([]int, nPerRow),
	}
}

func (g *simpleGrid) SetFilling(row int, column int) {
	if row < 0 || column < 0 || row > len(g.filling) || column > len(g.filling[row]) {
		panic(fmt.Errorf("coordinate (%d,%d) out of range passed to SimpleGrid.SetFilling()", row, column))
	}
	g.filling[row][column] = true
}

func (g *simpleGrid) SetStretchy(row int, column int) {
	if row < 0 || column < 0 || row > len(g.filling) || column > len(g.filling[row]) {
		panic(fmt.Errorf("coordinate (%d,%d) out of range passed to SimpleGrid.SetStretchy()", row, column))
	}
	if g.stretchyrow != -1 || g.stretchycol != -1 {
		g.filling[g.stretchyrow][g.stretchycol] = g.stretchyfill
	}
	g.stretchyrow = row
	g.stretchycol = column
	g.stretchyfill = g.filling[g.stretchyrow][g.stretchycol] // save previous value in case it changes later
	g.filling[g.stretchyrow][g.stretchycol] = true
}

func (g *simpleGrid) setParent(parent *controlParent) {
	for _, col := range g.controls {
		for _, c := range col {
			c.setParent(parent)
		}
	}
}

func (g *simpleGrid) allocate(x int, y int, width int, height int, d *sizing) (allocations []*allocation) {
	max := func(a int, b int) int {
		if a > b {
			return a
		}
		return b
	}

	var current *allocation // for neighboring

	if len(g.controls) == 0 {
		return nil
	}
	// 0) inset the available rect by the needed padding
	width -= (len(g.colwidths) - 1) * d.xpadding
	height -= (len(g.rowheights) - 1) * d.ypadding
	// 1) clear data structures
	for i := range g.rowheights {
		g.rowheights[i] = 0
	}
	for i := range g.colwidths {
		g.colwidths[i] = 0
	}
	// 2) get preferred sizes; compute row/column sizes
	for row, xcol := range g.controls {
		for col, c := range xcol {
			w, h := c.preferredSize(d)
			g.widths[row][col] = w
			g.heights[row][col] = h
			g.rowheights[row] = max(g.rowheights[row], h)
			g.colwidths[col] = max(g.colwidths[col], w)
		}
	}
	// 3) handle the stretchy control
	if g.stretchyrow != -1 && g.stretchycol != -1 {
		for i, w := range g.colwidths {
			if i != g.stretchycol {
				width -= w
			}
		}
		for i, h := range g.rowheights {
			if i != g.stretchyrow {
				height -= h
			}
		}
		g.colwidths[g.stretchycol] = width
		g.rowheights[g.stretchyrow] = height
	}
	// 4) draw
	startx := x
	for row, xcol := range g.controls {
		current = nil // reset on new columns
		for col, c := range xcol {
			w := g.widths[row][col]
			h := g.heights[row][col]
			if g.filling[row][col] {
				w = g.colwidths[col]
				h = g.rowheights[row]
			}
			as := c.allocate(x, y, w, h, d)
			if current != nil { // connect first left to first right
				current.neighbor = c
			}
			if len(as) != 0 {
				current = as[0] // next left is first subwidget
			} else {
				current = nil // spaces don't have allocation data
			}
			allocations = append(allocations, as...)
			x += g.colwidths[col] + d.xpadding
		}
		x = startx
		y += g.rowheights[row] + d.ypadding
	}
	return
}

// filling and stretchy are ignored for preferred size calculation
func (g *simpleGrid) preferredSize(d *sizing) (width int, height int) {
	max := func(a int, b int) int {
		if a > b {
			return a
		}
		return b
	}

	width -= (len(g.colwidths) - 1) * d.xpadding
	height -= (len(g.rowheights) - 1) * d.ypadding
	// 1) clear data structures
	for i := range g.rowheights {
		g.rowheights[i] = 0
	}
	for i := range g.colwidths {
		g.colwidths[i] = 0
	}
	// 2) get preferred sizes; compute row/column sizes
	for row, xcol := range g.controls {
		for col, c := range xcol {
			w, h := c.preferredSize(d)
			g.widths[row][col] = w
			g.heights[row][col] = h
			g.rowheights[row] = max(g.rowheights[row], h)
			g.colwidths[col] = max(g.colwidths[col], w)
		}
	}
	// 3) now compute
	for _, w := range g.colwidths {
		width += w
	}
	for _, h := range g.rowheights {
		height += h
	}
	return width, height
}

func (g *simpleGrid) commitResize(c *allocation, d *sizing) {
	// this is to satisfy Control; nothing to do here
}

func (g *simpleGrid) getAuxResizeInfo(d *sizing) {
	// this is to satisfy Control; nothing to do here
}
