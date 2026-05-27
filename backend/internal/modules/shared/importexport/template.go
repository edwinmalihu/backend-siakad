package importexport

import (
	"github.com/xuri/excelize/v2"
)

var studentsHeaders = []string{
	"NIS",
	"NISN",
	"Nama Lengkap",
	"Jenis Kelamin (male/female)",
	"Tempat Lahir",
	"Tanggal Lahir (YYYY-MM-DD)",
	"Alamat",
	"No. HP",
	"Tahun Masuk",
	"Status (active/inactive)",
}

var studentsSampleRow = []string{
	"2026001",
	"1234567890",
	"Andi Saputra",
	"male",
	"Padang",
	"2010-05-15",
	"Jl. Merdeka No. 10",
	"081234567890",
	"2026",
	"active",
}

func generateStudentsTemplate() ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Template"
	f.SetSheetName("Sheet1", sheet)

	for col, header := range studentsHeaders {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = f.SetCellValue(sheet, cell, header)
	}

	for col, val := range studentsSampleRow {
		cell, _ := excelize.CoordinatesToCellName(col+1, 2)
		_ = f.SetCellValue(sheet, cell, val)
	}

	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err == nil {
		_ = f.SetRowStyle(sheet, 1, 1, style)
	}

	_ = f.SetColWidth(sheet, "A", "A", 15)
	_ = f.SetColWidth(sheet, "B", "B", 15)
	_ = f.SetColWidth(sheet, "C", "C", 25)
	_ = f.SetColWidth(sheet, "D", "D", 22)
	_ = f.SetColWidth(sheet, "E", "E", 18)
	_ = f.SetColWidth(sheet, "F", "F", 22)
	_ = f.SetColWidth(sheet, "G", "G", 30)
	_ = f.SetColWidth(sheet, "H", "H", 18)
	_ = f.SetColWidth(sheet, "I", "I", 14)
	_ = f.SetColWidth(sheet, "J", "J", 18)

	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomRight",
	})

	genderDV := excelize.NewDataValidation(true)
	genderDV.SetSqref("D2:D1048576")
	_ = genderDV.SetDropList([]string{"male", "female"})
	_ = f.AddDataValidation(sheet, genderDV)

	statusDV := excelize.NewDataValidation(true)
	statusDV.SetSqref("J2:J1048576")
	_ = statusDV.SetDropList([]string{"active", "inactive"})
	_ = f.AddDataValidation(sheet, statusDV)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
