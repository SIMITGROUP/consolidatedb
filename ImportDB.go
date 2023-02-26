package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func ImportCSVToDB(foldername string, tables []string) {

	for _, table := range tables {

		filename := filepath.Join(datafolder, foldername, table+".csv")
		file, err := os.Open(filename)
		if err != nil {
			logrus.Fatal(err.Error())
		}
		reader := csv.NewReader(file)
		data, err2 := reader.ReadAll()
		columns := []string{}

		if err2 == nil {
			for i, row := range data {
				if i == 0 {
					columns = row
					logrus.Info(columns)
				} else {
					logrus.Info(row)
				}
			}

		} else {
			logrus.Fatal(err2)
		}
		logrus.Info("Import data ", filename)
		sql := fmt.Sprintf("LOAD DATA INFILE '%s' INTO TABLE %s FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\n' IGNORE 1 ROWS", filename, table)
		logrus.Warn(sql)
		// res, err := localdb.Query(sql)
		res, err := localdb.Exec(sql)
		logrus.Info(res, err)

	}

	// db, err := ConnectDB(dbsetting)
	// sql := "SELECT * FROM ?"
	// if err == nil {
	// 	for _, tablename := range tables {
	// 		res, errquery := db.Query(sql, tablename)
	// 		if errquery == nil {

	// 		} else {
	// 			logrus.Fatal(errquery)
	// 		}

	// 	}

	// }
}

func ImportData(dbsetting Model_DBSetting, tables []string) (err error) {
	//connect to specific tenant
	db, err := ConnectDB(dbsetting)

	//loop through all table
	for _, tablename := range tables {
		// get all data
		sql := fmt.Sprintf("SELECT %s FROM %s", mapfieldstr[tablename], tablename)
		logrus.Info(sql)
		rows, err := db.Query(sql)
		if err == nil {
			//insert all data into local db
			err = InsertRecord(dbsetting.Tenant_id, tablename, rows)

		} else {
			logrus.Fatal(err)
		}
	}

	return
}

func InsertRecord(tenant_id string, tablename string, rows *sql.Rows) (err error) {
	// seperator := []byte("\t")
	cols := []string{}
	oricolumns, _ := rows.Columns()
	for _, colname := range oricolumns {
		if colname != "tenant_id" {
			cols = append(cols, colname)
		}
	}

	// colstring := strings.Join(cols, ",")
	row := make([][]byte, len(cols))
	// row[0] = []byte(tenant_id)
	rowPtr := make([]any, len(cols))
	for i := range row {
		rowPtr[i] = &row[i]
	}
	colstring := "`tenant_id`, " + mapfieldstr[tablename]
	// rowstring := "'" + tenant_id + "'"
	// for _, f := range cols {
	// 	numeric_precision, found := mapfields[tablename].Get(f)
	// 	if found == true {
	// 		valustr := "%v"
	// 		if numeric_precision == false {
	// 			valustr = "%s"
	// 		}

	// 		rowstring += "," + valustr
	// 	} else {
	// 		logrus.Fatal("Field '", f, "' does not exists in table '", tablename, "'")
	// 	}
	// }

	for rows.Next() {
		err = rows.Scan(rowPtr...)
		tmp := "'" + tenant_id + "'"
		for i, f := range cols {
			value := string(row[i])

			numeric_precision, found := mapfields[tablename].Get(f)
			if found == true && numeric_precision == true {
				logrus.Warn(f, ": before ", fmt.Sprintf("%#v", value))
				if value == "" {
					value = "0"
				}
				logrus.Warn(f, ": after ", fmt.Sprintf("%#v", value))
				tmp += ", " + value
			} else {
				tmp += ", '" + MysqlRealEscapeString(value) + "'"
			}

		}
		// rowstring := string(bytes.Join(row, seperator))

		sqlstr := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tablename, colstring, tmp)
		_, errinsert := localdb.Exec(sqlstr)
		logrus.Info(sqlstr)

		// _, errinsert := local(sqlstr, row)

		if errinsert != nil {
			logrus.Fatal(errinsert)
		}
	}

	return
}

func MysqlRealEscapeString(value string) string {
	var sb strings.Builder
	for i := 0; i < len(value); i++ {
		c := value[i]
		switch c {
		case '\\', 0, '\n', '\r', '\'', '"':
			sb.WriteByte('\\')
			sb.WriteByte(c)
		case '\032':
			sb.WriteByte('\\')
			sb.WriteByte('Z')
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}
