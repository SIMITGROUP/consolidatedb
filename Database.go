package main

import (
	"database/sql"
	"fmt"

	"github.com/blockloop/scan"
	"github.com/emirpasic/gods/maps/hashmap"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

func ConnectLocalDB() (*sql.DB, error) {

	// host string, dbname string, user string, pass string
	dbsetting := Model_DBSetting{
		Host: localdbhost,
		Db:   localdbname,

		User: localdbuser,
		Pass: localdbpass,
	}
	return ConnectDB(dbsetting)
}

func ConnectDB(dbsetting Model_DBSetting) (*sql.DB, error) {

	connstr := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbsetting.User, dbsetting.Pass, dbsetting.Host, dbsetting.Db)
	db, err := sql.Open("mysql", connstr)
	if err != nil {
		return nil, err
	} else {
		return db, nil
	}

}

//	func GetLocalTables() []string {
//		return GetAllTables(localdb, localdbname)
//	}
func GetAllTables(db *sql.DB, dbname string) []string {
	sql_tablelist := "SELECT table_name FROM INFORMATION_SCHEMA.TABLES WHERE table_type = 'BASE TABLE' and table_schema=?"

	var listtable = []string{}

	res, err := db.Query(sql_tablelist, dbname)

	if err == nil {

		for res.Next() {
			var tablename string

			res.Scan(&tablename)

			mapfields[tablename] = hashmap.New()
			listtable = append(listtable, tablename)
		}
	} else {

		logrus.Fatal(err)
	}

	return listtable

}
func GetAllTableAndFields(db *sql.DB, dbname string) {

	sql_tablelist := "select table_name,column_name,column_type, data_type,column_default, is_nullable, column_key,extra,character_set_name,collation_name,numeric_precision from information_schema.columns where table_schema = ? order by table_name,ordinal_position"
	rows, err := db.Query(sql_tablelist, dbname)
	// sql_select := "SELECT %s FROM %s"
	logrus.Info("SQL:", sql_tablelist)
	if err == nil {
		i := 0
		for rows.Next() {
			i = i + 1
			tablename := ""
			f := Model_FieldSetting{}

			rows.Scan(&tablename, &f.Field, &f.Type, &f.DataType, &f.Default, &f.Null, &f.Key, &f.Extra, &f.CharacterSet, &f.CollationName, &f.NumericPrecision)
			//exclude existing tenant field, so we can manage it ourself during import
			if f.Field == "tenant_id" {
				continue
			} else {
				fields := mapfields[tablename] //mapfields.Get(tablename)
				// logrus.Warn("add ", tablename, ":", f.Field, ":", f.NumericPrecision.String)
				ftype := "string"
				if f.NumericPrecision.Valid == true {
					ftype = "number"
				} else if f.DataType == "date" || f.DataType == "datetime" {
					ftype = f.DataType
				}

				fields.Put(f.Field, ftype)
			}
		}

	}

	for tablename, fields := range mapfields {
		// logrus.Error(tablename, fields.Keys())
		for i, fieldname := range fields.Keys() {
			if i > 0 {
				mapfieldstr[tablename] += ", "
			}

			mapfieldstr[tablename] += fmt.Sprintf("`%v`", fieldname)
		}
	}
}

// parse csv columns, create query statement
func parseColumns(TABLENAME string, columns []string, query *string) {

	*query = "INSERT INTO " + TABLENAME + " ("
	placeholder := "VALUES ("
	for i, c := range columns {
		if i == 0 {
			*query += c
			placeholder += "?"
		} else {
			*query += ", " + c
			placeholder += ", ?"
		}
	}
	placeholder += ")"
	*query += ") " + placeholder
}

// convert []string to []interface{}
func string2interface(s []string) []interface{} {

	i := make([]interface{}, len(s))
	for k, v := range s {
		i[k] = v
	}
	return i
}

func GetRemoteDatabases() (dbsettings []Model_DBSetting) {
	sql := "SELECT * FROM tenant_master where isactive=1"
	if RunMode == "append" {
		sql += " and imported ='' "
	}
	// logrus.Fatal("Ping", sql, localdb)

	res, err := localdb.Query(sql)

	if err == nil {
		err = scan.Rows(&dbsettings, res)
		if err == nil {
			return
		} else {
			logrus.Fatal(err)
		}
	} else {
		logrus.Fatal(err)
	}
	return

}
