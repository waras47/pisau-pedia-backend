package export

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// ToExcel renders a Table into a single-sheet .xlsx file: title merged
// across the header row, bold column headers, then one row per data row.
func ToExcel(t Table) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	const sheet = "Sheet1"

	if len(t.Headers) > 0 {
		lastCol, err := excelize.ColumnNumberToName(len(t.Headers))
		if err != nil {
			return nil, err
		}

		titleStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}})
		if err != nil {
			return nil, err
		}
		if err := f.SetCellValue(sheet, "A1", t.Title); err != nil {
			return nil, err
		}
		if err := f.MergeCell(sheet, "A1", fmt.Sprintf("%s1", lastCol)); err != nil {
			return nil, err
		}
		if err := f.SetCellStyle(sheet, "A1", "A1", titleStyle); err != nil {
			return nil, err
		}

		headerStyle, err := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#E5E7EB"}, Pattern: 1},
		})
		if err != nil {
			return nil, err
		}
		for i, h := range t.Headers {
			cell, err := excelize.CoordinatesToCellName(i+1, 3)
			if err != nil {
				return nil, err
			}
			if err := f.SetCellValue(sheet, cell, h); err != nil {
				return nil, err
			}
			if err := f.SetCellStyle(sheet, cell, cell, headerStyle); err != nil {
				return nil, err
			}
		}
	}

	for rowIdx, row := range t.Rows {
		for colIdx, val := range row {
			cell, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+4)
			if err != nil {
				return nil, err
			}
			if err := f.SetCellValue(sheet, cell, val); err != nil {
				return nil, err
			}
		}
	}

	for i := range t.Headers {
		col, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			return nil, err
		}
		if err := f.SetColWidth(sheet, col, col, 22); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
