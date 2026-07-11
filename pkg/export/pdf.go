package export

import (
	"bytes"

	"github.com/go-pdf/fpdf"
)

// ToPDF renders a Table as a simple landscape A4 document: title, then a
// bordered table with a shaded header row. Column widths are split evenly
// across the printable width — good enough for report-style tabular data,
// not meant for pixel-perfect layouts.
func ToPDF(t Table) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(10, 15, 10)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, t.Title, "", 1, "L", false, 0, "")
	pdf.Ln(4)

	if len(t.Headers) == 0 {
		buf := new(bytes.Buffer)
		if err := pdf.Output(buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	pageWidth, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	usableWidth := pageWidth - left - right
	colWidth := usableWidth / float64(len(t.Headers))
	rowHeight := 8.0

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(229, 231, 235)
	for _, h := range t.Headers {
		pdf.CellFormat(colWidth, rowHeight, h, "1", 0, "L", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 9)
	for _, row := range t.Rows {
		for i, val := range row {
			if i >= len(t.Headers) {
				break
			}
			pdf.CellFormat(colWidth, rowHeight, val, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}

	buf := new(bytes.Buffer)
	if err := pdf.Output(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
