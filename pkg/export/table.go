package export

// Table is a generic tabular dataset shared by the Excel and PDF generators,
// so a report only needs to build this once and can be exported as either
// format without duplicating layout logic.
type Table struct {
	Title   string
	Headers []string
	Rows    [][]string
}
