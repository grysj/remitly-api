package parser

import (
	"encoding/csv"
	"fmt"
	"os"
)

func ParseCSV(pathToCSV string) ([]CsvRow, error) {

	file, err := os.Open(pathToCSV)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	idxMap, err := createColumnMap(header)
	if err != nil {
		return nil, err
	}

	var records []CsvRow

	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		column := CsvRow{
			ISO2:     row[idxMap["COUNTRY ISO2 CODE"]],
			Swift:    row[idxMap["SWIFT CODE"]],
			Type:     row[idxMap["CODE TYPE"]],
			Name:     row[idxMap["NAME"]],
			Address:  row[idxMap["ADDRESS"]],
			Town:     row[idxMap["TOWN NAME"]],
			Country:  row[idxMap["COUNTRY NAME"]],
			Timezone: row[idxMap["TIME ZONE"]],
		}
		records = append(records, column)
	}

	return records, nil

}

func getColumnIdx(header []string, columnName string) int {

	for i, column := range header {
		if columnName == column {
			return i
		}
	}
	return -1
}

func createColumnMap(header []string) (map[string]int, error) {

	requiredColumns := []string{"COUNTRY ISO2 CODE", "SWIFT CODE", "NAME", "TOWN NAME", "COUNTRY NAME", "TIME ZONE", "CODE TYPE", "ADDRESS"}
	indexMap := make(map[string]int)

	for _, req := range requiredColumns {
		idx := getColumnIdx(header, req)
		if idx < 0 {
			return nil, fmt.Errorf("Error: %s column not found in CSV", req)
		}
		indexMap[req] = idx
	}

	return indexMap, nil
}
