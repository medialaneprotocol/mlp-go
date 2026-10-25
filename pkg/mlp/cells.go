package mlp

// CellGrid describes a media surface geometry.
type CellGrid struct {
	WidthPx  uint32
	HeightPx uint32
	CellSize uint32
}

// Cols is the grid's column count.
func (g CellGrid) Cols() uint32 {
	return (g.WidthPx + g.CellSize - 1) / g.CellSize
}

// Rows is the grid's row count.
func (g CellGrid) Rows() uint32 {
	return (g.HeightPx + g.CellSize - 1) / g.CellSize
}

// Cells is the grid's total cell count.
func (g CellGrid) Cells() uint32 { return g.Cols() * g.Rows() }

// CellAccounting tracks per-receiver convergence over a repair-cell grid:
// a cell is pending from the moment an update covering it is sent until the
// receiver applies it.
type CellAccounting struct {
	Grid    CellGrid
	Pending map[uint32]bool
	Applied map[uint32]bool
}

func NewCellAccounting(grid CellGrid) *CellAccounting {
	return &CellAccounting{Grid: grid, Pending: map[uint32]bool{}, Applied: map[uint32]bool{}}
}

// MarkPending declares cells sent-but-unapplied.
func (a *CellAccounting) MarkPending(cells []uint32) {
	for _, c := range cells {
		a.Pending[c] = true
		delete(a.Applied, c)
	}
}

// MarkApplied declares cells applied by the receiver.
func (a *CellAccounting) MarkApplied(cells []uint32) {
	for _, c := range cells {
		delete(a.Pending, c)
		a.Applied[c] = true
	}
}

// PendingCount is the number of unapplied cells.
func (a *CellAccounting) PendingCount() int { return len(a.Pending) }

// Converged reports whether nothing is pending.
func (a *CellAccounting) Converged() bool { return len(a.Pending) == 0 }
