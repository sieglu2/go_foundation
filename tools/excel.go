package foundation

import (
	"fmt"
	"reflect"
	"time"

	"github.com/sieglu2/go_foundation/foundation"
	"github.com/xuri/excelize/v2"
)

var (
	ErrNoDataToExport  = fmt.Errorf("no data to export")
	ErrEmptyFileHandle = fmt.Errorf("empty excel file handle")
	ErrNilDataProvided = fmt.Errorf("nil data provided")
)

type ExcelWriter struct {
	fileHandle *excelize.File
}

func CreateEmptyExcelSheet(fileName string, sheetNames []string) (*ExcelWriter, error) {
	logger := foundation.Logger()

	if len(sheetNames) == 0 {
		logger.Errorf("no sheetNames given")
		return nil, fmt.Errorf("no sheetNames given")
	}

	fileHandle := excelize.NewFile()

	defaultSheet := fileHandle.GetSheetName(0)
	if err := fileHandle.SetSheetName(defaultSheet, sheetNames[0]); err != nil {
		logger.Errorf("failed to rename default sheet: %v", err)
		return nil, err
	}

	for i := 1; i < len(sheetNames); i += 1 {
		if _, err := fileHandle.NewSheet(sheetNames[i]); err != nil {
			logger.Errorf("failed to create sheet (%s): %v", sheetNames[i], err)
			return nil, err
		}
	}

	// Save the file
	if err := fileHandle.SaveAs(fileName); err != nil {
		logger.Errorf("failed to save excel file: %v", err)
		return nil, err
	}

	return &ExcelWriter{
		fileHandle: fileHandle,
	}, nil
}

func (c *ExcelWriter) SaveFile() error {
	logger := Logger()

	if c.fileHandle == nil {
		logger.Warn(ErrEmptyFileHandle)
		return ErrEmptyFileHandle
	}

	if err := c.fileHandle.Save(); err != nil {
		logger.Errorf("failed to save excel file: %v", err)
		return err
	}

	return nil
}

func (c *ExcelWriter) AppendDataAsRows(sheetName string, labelTagStr string, bars []any) error {
	logger := Logger()

	if c.fileHandle == nil {
		logger.Warn(ErrEmptyFileHandle)
		return ErrEmptyFileHandle
	}

	if len(bars) == 0 {
		logger.Warn(ErrNoDataToExport)
		return ErrNoDataToExport
	}

	// Get the current row count to know where to start appending
	rows, err := c.fileHandle.GetRows(sheetName)
	if err != nil {
		logger.Errorf("failed to get rows: %v", err)
		return err
	}

	startRow := len(rows)
	if startRow == 0 {
		// If sheet is empty, write headers first
		headers := getHeadersFromTags(labelTagStr, bars[0])
		for col, header := range headers {
			cell := fmt.Sprintf("%c1", 'A'+col)
			if err := c.fileHandle.SetCellValue(sheetName, cell, header); err != nil {
				logger.Errorf("failed to write header: %v", err)
				return err
			}
		}
		startRow = 1
	}

	// Write data rows
	for i, bar := range bars {
		rowNum := startRow + i + 1 // Start from next row after existing data
		values := getValuesFromStruct(bar)

		for col, value := range values {
			cell := fmt.Sprintf("%c%d", 'A'+col, rowNum)
			if err := c.fileHandle.SetCellValue(sheetName, cell, value); err != nil {
				logger.Errorf("failed to write value at %s: %v", cell, err)
				return err
			}
		}
	}

	return nil
}

func (c *ExcelWriter) WriteStructFieldsAsRows(sheetName string, labelTagStr string, data any) error {
	logger := Logger()

	if c.fileHandle == nil {
		logger.Warn(ErrEmptyFileHandle)
		return ErrEmptyFileHandle
	}

	if data == nil {
		logger.Error(ErrNilDataProvided)
		return ErrNilDataProvided
	}

	// Get the value that the pointer points to
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem() // Dereference the pointer
	}

	// Check if the dereferenced value is a struct
	if v.Kind() != reflect.Struct {
		logger.Errorf("data must be a struct or pointer to struct")
		return fmt.Errorf("data must be a struct or pointer to struct")
	}

	// Get the type information of the struct
	t := v.Type()

	// Write each field as a row
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Get the label tag if it exists
		labelTag := field.Tag.Get(labelTagStr)
		fieldName := labelTag
		if fieldName == "" {
			fieldName = field.Name
		}

		rowNum := i + 1

		// Write field name
		cellField := fmt.Sprintf("A%d", rowNum)
		if err := c.fileHandle.SetCellValue(sheetName, cellField, fieldName); err != nil {
			logger.Errorf("failed to write field name at %s: %v", cellField, err)
			return err
		}

		// Handle slice fields differently
		if value.Kind() == reflect.Slice {
			for j := 0; j < value.Len(); j++ {
				// Convert column number to Excel column letter (B, C, D, etc.)
				colLetter := string(rune('B' + j))
				cellValue := fmt.Sprintf("%s%d", colLetter, rowNum)

				sliceValue := value.Index(j).Interface()
				if err := c.fileHandle.SetCellValue(sheetName, cellValue, sliceValue); err != nil {
					logger.Errorf("failed to write slice value at %s: %v", cellValue, err)
					return err
				}
			}
		} else {
			// Write non-slice value
			cellValue := fmt.Sprintf("B%d", rowNum)
			if err := c.fileHandle.SetCellValue(sheetName, cellValue, value.Interface()); err != nil {
				logger.Errorf("failed to write value at %s: %v", cellValue, err)
				return err
			}
		}
	}

	return nil
}

func getHeadersFromTags(labelTagStr string, c any) []string {
	t := reflect.TypeOf(c)
	var headers []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if header, ok := field.Tag.Lookup(labelTagStr); ok {
			headers = append(headers, header)
		} else {
			headers = append(headers, field.Name)
		}
	}

	return headers
}

func getValuesFromStruct(c any) []interface{} {
	v := reflect.ValueOf(c)
	t := v.Type()
	var values []interface{}

	// Load ET timezone once
	et, err := time.LoadLocation("America/New_York")
	if err != nil {
		et = time.Local // fallback to local timezone if ET not available
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		switch field.Interface().(type) {
		case time.Time:
			timeVal := field.Interface().(time.Time)
			// Convert to ET and format with timezone
			timeInET := timeVal.In(et)
			if format, ok := t.Field(i).Tag.Lookup("format"); ok {
				values = append(values, timeInET.Format(format+" MST"))
			} else {
				values = append(values, timeInET.Format("20060102 15:04 MST"))
			}
		default:
			values = append(values, field.Interface())
		}
	}

	return values
}
