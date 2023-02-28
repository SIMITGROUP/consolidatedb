package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// func ImportCSVToDB(foldername string, tables []string) {

// 	for _, table := range tables {

// 		filename := filepath.Join(datafolder, foldername, table+".csv")
// 		file, err := os.Open(filename)
// 		if err != nil {
// 			logrus.Fatal(err.Error())
// 		}
// 		reader := csv.NewReader(file)
// 		data, err2 := reader.ReadAll()
// 		columns := []string{}

// 		if err2 == nil {
// 			for i, row := range data {
// 				if i == 0 {
// 					columns = row
// 					logrus.Info(columns)
// 				} else {
// 					logrus.Info(row)
// 				}
// 			}

// 		} else {
// 			logrus.Fatal(err2)
// 		}
// 		logrus.Info("Import data ", filename)
// 		sql := fmt.Sprintf("LOAD DATA INFILE '%s' INTO TABLE %s FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\n' IGNORE 1 ROWS", filename, table)
// 		logrus.Warn(sql)
// 		// res, err := localdb.Query(sql)
// 		res, err := localdb.Exec(sql)
// 		logrus.Info(res, err)

// 	}

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
// }

func ImportData(dbsetting Model_DBSetting, tables []string) (err error) {
	//connect to specific tenant
	db, err := ConnectDB(dbsetting)
	internaldb, _ := ConnectDB(localdbsetting)
	//loop through all table
	// tables1 := []string{"acc_payment"}
	for _, tablename := range tables {
		// get all data
		exists, _ := in_array(tablename, excludedtables)
		if exists {
			continue
		}
		sql := fmt.Sprintf("SELECT %s FROM %s", mapfieldstr[tablename], tablename)
		// logrus.Info(sql)
		rows, err := db.Query(sql)
		if err == nil {
			//insert all data into local db
			err = InsertRecord(dbsetting.Tenant_id, internaldb, tablename, rows)
			rows.Close()
		} else {

			logrus.Fatal(err)
		}
	}
	currentTime := time.Now()

	internaldb.Exec("Update tenant_master set imported= ? WHERE tenant_id=?", dbsetting.Tenant_id, currentTime.Format("2006.01.02 15:04:05"))
	return
}

func InsertRecord(tenant_id string, internaldb *sql.DB, tablename string, rows *sql.Rows) (err error) {
	// seperator := []byte("\t")
	cols := []string{}
	oricolumns, _ := rows.Columns()
	for _, colname := range oricolumns {
		if colname != "tenant_id" {
			cols = append(cols, colname)
		}
	}
	fmt.Print("Import [", tenant_id, "] => ", tablename, " x ")
	// colstring := strings.Join(cols, ",")
	row := make([][]byte, len(cols))
	// row[0] = []byte(tenant_id)
	rowPtr := make([]interface{}, len(cols))
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
	rowcount := 0
	batchsize := 50
	batchno := 0
	rowstring := ""
	for rows.Next() {
		rowcount++
		batchno++
		err = rows.Scan(rowPtr...)
		tmp := "'" + tenant_id + "'"
		for i, f := range cols {
			value := string(row[i])

			fieldtype, found := mapfields[tablename].Get(f)
			if found == true {

				if fieldtype == "number" {
					if value == "" {
						value = "0"
					}
					tmp += ", " + value
				} else if fieldtype == "date" {
					if value == "" {
						value = "0000-00-00"
					}
					tmp += ", '" + value + "'"

				} else if fieldtype == "datetime" {
					if value == "" {
						value = "0000-00-00 00:00:00"
					}
					if f == "document_trackupdated" && value != "0000-00-00 00:00:00" {
						logrus.Warn(tenant_id, "=> ", row[50], " => document_trackupdated: ", value)
					}
					tmp += ", '" + value + "'"

				} else {
					tmp += ", '" + MysqlRealEscapeString(value) + "'"
				}

			}

		}

		if batchno >= batchsize {
			batchno = 0
			rowstring += ",(" + tmp + ")"
			sqlstr := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tablename, colstring, rowstring)
			rowstring = ""
			// logrus.Warn(sqlstr)
			_, errinsert := internaldb.Exec(sqlstr)
			if errinsert != nil {
				logrus.Error(errinsert)
				logrus.Fatal(sqlstr)
			}
		} else {
			if rowstring == "" {
				rowstring = "(" + tmp + ")"
			} else {
				rowstring += ",(" + tmp + ")"
			}
			// logrus.Warn(batchno, " : ", tmp)

		}
		// rowstring := string(bytes.Join(row, seperator))

	}

	if rowstring != "" {
		sqlstr := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tablename, colstring, rowstring)
		_, errinsert := internaldb.Exec(sqlstr)
		if errinsert != nil {

			logrus.Fatal(tablename, errinsert)
		}
	}
	fmt.Println(rowcount)
	// if rowcount > 0 {
	// 	log.Fatal("Done")
	// }

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

func in_array(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}
