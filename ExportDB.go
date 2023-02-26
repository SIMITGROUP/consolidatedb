package main

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/joho/sqltocsv"
	"github.com/sirupsen/logrus"
)

func ExportDBToCSV(dbsetting Model_DBSetting, tables []string) {
	db, err := ConnectDB(dbsetting)
	path := filepath.Join(datafolder, dbsetting.Tenant_id)
	CreateFolderIfNotExists(path)

	if err == nil {
		for _, tablename := range tables {
			logrus.Info(tablename)
			//ignore tenantmas
			if tablename != table_tenant {
				ExportTableToCSV(db, path, tablename)
			}

		}
	}

}

func ExportTableToCSV(db *sql.DB, path string, tablename string) {
	sql := fmt.Sprintf("SELECT * FROM %s", tablename)
	logrus.Info(sql)
	rows, err := db.Query(sql)
	if err == nil {

		filename := filepath.Join(path, tablename+".csv")

		err = sqltocsv.WriteFile(filename, rows)
		if err == nil {
			logrus.Info("Exported ", tablename)
		} else {
			logrus.Fatal(err)
		}
	} else {
		logrus.Fatal(err)
	}

}

// func WriteFile(csvFileName string, rows *sql.Rows) error {
// 	return New(rows).WriteFile(csvFileName)
// }

// // New will return a Converter which will write your CSV however you like
// // but will allow you to set a bunch of non-default behaivour like overriding
// // headers or injecting a pre-processing step into your conversion
// func New(rows *sql.Rows) *Converter {
// 	return &Converter{
// 		rows:         rows,
// 		WriteHeaders: true,
// 		Delimiter:    ',',
// 	}
// }
