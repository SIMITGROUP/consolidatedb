package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var localdb *sql.DB
var table_tenant string = "tenant_master"
var localdbname string = ""
var localdbhost string = ""
var localdbuser string = ""
var localdbpass string = ""
var datafolder string = ""
var mastertables []string
var Delimiter = ','
var mapfields = make(map[string]*hashmap.Map)
var mapfieldstr = make(map[string]string)

func main() {
	// sql_tablelist := "SELECT table_name FROM INFORMATION_SCHEMA.TABLES WHERE table_type = 'BASE TABLE' and table_schema=?"
	var wg sync.WaitGroup
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cwd, _ := os.Getwd()
	datafolder = cwd + "/data"
	localdbname = os.Getenv("dbname")
	localdbhost = os.Getenv("dbhost")
	localdbuser = os.Getenv("dbuser")
	localdbpass = os.Getenv("dbpass")

	fmt.Println("Welcome mysql db consoler:")

	CreateFolderIfNotExists(datafolder)

	localdb, err = ConnectLocalDB()

	defer localdb.Close()

	if err == nil {
		fmt.Println("connected")

		localtables := GetAllTables(localdb, localdbname) // GetLocalTables()

		// names := lo.Uniq[string]([]string{"Samuel", "John", "Samuel"})
		tablecount := len(localtables)
		logrus.Info("table count at local:", tablecount)

		if tablecount == 0 { //
			logrus.Fatal(localdbname + " does not have table tenant_master")
		}

		if tablecount == 1 {
			//run create tables
		}
		// logrus.Fatal("GetRemoteDatabases")
		dbsettings := GetRemoteDatabases()

		if len(dbsettings) == 0 {
			logrus.Fatal("no record found in tenant_master table")
		} else {
			mastertables = GenerateTables(dbsettings[0])

		}
		for i, setting := range dbsettings {
			wg.Add(1) // declare new go routine added
			go func(i int, dbsetting Model_DBSetting, tables []string) {
				logrus.Info(i, dbsetting)
				// ExportDBToCSV(dbsetting, tables)
				// ImportCSVToDB(dbsetting.Tenant_id, tables)
				localdb.Exec("Set @@SQL_MODE=''")
				ImportData(dbsetting, tables)
				defer wg.Done()
			}(i, setting, mastertables)
		}
		logrus.Info("before wait")
		wg.Wait()
		logrus.Info("after wait")
		// for _, tablename := range localtables {
		// 	//play safe, exclude tenant_master table
		// 	if tablename != table_tenant {

		// 	}
		// }

	} else {
		logrus.Fatal("connect failed")
	}

}

// generate master tables according first db setting, return list of tables
func GenerateTables(dbsetting Model_DBSetting) (tables []string) {
	db, err := ConnectDB(dbsetting)
	if err == nil {
		tables = GetAllTables(db, dbsetting.Db)
		GetAllTableAndFields(db, dbsetting.Db)
		for i, j := range mapfields {
			logrus.Info(i, ":", j.Keys())
		}
	}
	/*index data
	[{
		indexname: index1,
		indextype: PRI/INDEX/Unique
		columns: [col1,col2,col3]

	},
	{}]
	*/
	for _, tablename := range tables {
		//drop local table if exists
		dropsql := "DROP TABLE IF EXISTS " + tablename
		_, errdrop := localdb.Query(dropsql)
		if errdrop != nil {
			logrus.Fatal(errdrop)
		}
		var tablesetting []Model_FieldSetting

		sql := "DESCRIBE " + tablename
		logrus.Info(sql)
		rows, err2 := db.Query(sql)
		if err2 == nil {
			var fieldsetting Model_FieldSetting
			var primarykey = ""
			for rows.Next() {
				err3 := rows.Scan(&fieldsetting.Field, &fieldsetting.Type, &fieldsetting.Null, &fieldsetting.Key, &fieldsetting.Default, &fieldsetting.Extra)
				if err3 != nil {
					logrus.Fatal(err3)
				}
				tablesetting = append(tablesetting, fieldsetting)
			}

			sqlcreate := "CREATE TABLE " + tablename + " (`tenant_id` varchar(20) "

			for _, s := range tablesetting {
				// ignore tenant_id field
				if s.Field == "tenant_id" {
					continue
				}
				logrus.Info(s.Field, ":", s.Key)
				if s.Key == "PRI" {
					primarykey = "tenant_id," + s.Field
				}

				sqlcreate += ", `" + s.Field + "` " + s.Type
			}

			if primarykey == "" {
				logrus.Fatal(tablename + " undefined primarykey")
			}
			sqlcreate = sqlcreate + ", PRIMARY KEY (" + primarykey + ")) ENGINE=InnoDB"
			logrus.Info(sqlcreate)
			_, errcreate := localdb.Query(sqlcreate)
			if errcreate != nil {
				logrus.Fatal(errcreate)
			}

		} else {
			logrus.Fatal(err2)
		}

	}
	db.Close()
	//logrus.Fatal("Generate tables end")
	return

}
func CreateFolderIfNotExists(path string) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

}
