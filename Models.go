package main

import "database/sql"

type CsvPreProcessorFunc func(row []string, columnNames []string) (outputRow bool, processedRow []string)

type Model_DBSetting struct {
	Tenant_id   string `json:"tenant_id"`
	Host        string `json:"host"`
	Db          string `json:"db"`
	User        string `json:"user"`
	Pass        string `json:"pass"`
	Description string `json:"description"`
	Imported    string `json:"imported"`
}

type Converter struct {
	Headers      []string // Column headers to use (default is rows.Columns())
	WriteHeaders bool     // Flag to output headers in your CSV (default is true)
	TimeFormat   string   // Format string for any time.Time values (default is time's default)
	FloatFormat  string   // Format string for any float64 and float32 values (default is %v)
	Delimiter    rune     // Delimiter to use in your CSV (default is comma)

	rows            *sql.Rows
	rowPreProcessor CsvPreProcessorFunc
}
type Model_TableIndex struct {
	Name      string
	IndexType string
	Columns   []string
}
type Model_FieldSetting struct {
	Field            string
	Type             string
	DataType         string
	Null             string
	Key              string
	Default          sql.NullString
	Extra            string
	CharacterSet     sql.NullString
	CollationName    sql.NullString
	NumericPrecision sql.NullString
}
