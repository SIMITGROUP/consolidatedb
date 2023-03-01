package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/emirpasic/gods/maps/hashmap"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var localdb *sql.DB
var table_tenant string = "tenant_master"

// var localdbname string = ""
// var localdbhost string = ""
// var localdbuser string = ""
// var localdbpass string = ""
var localdbsetting Model_DBSetting
var datafolder string = ""
var mastertables []string
var Delimiter = ','
var mapfields = make(map[string]*hashmap.Map)
var mapfieldstr = make(map[string]string)
var excludedtables = []string{"tenant_master", "gps_event", "system_event", "system_image", "cs_ticketline"}
var RunMode = ""

// const MAX_CONCURRENT_JOBS = 4

func main() {
	var wg sync.WaitGroup
	args := os.Args
	if len(args) < 2 {
		logrus.Fatal("Undefine run mode argument, supply argument 'init' or 'append'.")
	} else {
		RunMode = args[1]
	}
	err := godotenv.Load()
	start := time.Now()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	localdbsetting.Db = os.Getenv("dbname")
	localdbsetting.Host = os.Getenv("dbhost")
	localdbsetting.User = os.Getenv("dbuser")
	localdbsetting.Pass = os.Getenv("dbpass")
	localdb, err = ConnectDB(localdbsetting)
	localdb.SetMaxOpenConns(20)
	defer localdb.Close()
	var dbsettings []Model_DBSetting
	if err == nil {
		logrus.Info(localdbsetting.Db, " connected")

		if len(args) >= 2 {
			dbsettings = GetRemoteDatabases()
			if len(dbsettings) == 0 {
				logrus.Fatal("no tenant record ready for import")
			}

			if RunMode == "init" {
				mastertables = GetTables(dbsettings[0])
				err := GenerateTables(dbsettings[0], mastertables)
				if err != nil {
					logrus.Fatal(err)
				}

			} else if RunMode == "append" {
				mastertables = GetTables(localdbsetting)
			} else {
				logrus.Fatal("mode '", args[1], "' is not supported. Please supply argument 'init' or 'append'")
			}
		}

		//close connection. every import create new connection instead
		localdb.Close()
		// waitChan := make(chan struct{}, MAX_CONCURRENT_JOBS)
		// count := 0

		for i, setting := range dbsettings {
			// waitChan <- struct{}{}
			// count++

			wg.Add(1) // declare new go routine added
			go func(i int, dbsetting Model_DBSetting, tables []string) {
				logrus.Info(i, "import tenant :", dbsetting.Tenant_id)

				// ExportDBToCSV(dbsetting, tables)
				// ImportCSVToDB(dbsetting.Tenant_id, tables)

				ImportData(dbsetting, tables)
				// <-waitChan
				defer wg.Done()
			}(i, setting, mastertables)
		}
		// logrus.Info("before wait")
		wg.Wait()

		// for _, tablename := range localtables {
		// 	//play safe, exclude tenant_master table
		// 	if tablename != table_tenant {

		// 	}
		// }

	} else {
		logrus.Fatal("connect failed")
	}

	r := new(big.Int)
	fmt.Println(r.Binomial(1000, 10))

	elapsed := time.Since(start)
	log.Printf("Binomial took %s", elapsed)

}

// generate master tables according first db setting, return list of tables
func GetTables(dbsetting Model_DBSetting) (tables []string) {
	db, err := ConnectDB(dbsetting)
	if err == nil {
		tables = GetAllTables(db, dbsetting.Db)
		GetAllTableAndFields(db, dbsetting.Db)
		// for i, j := range mapfields {
		// 	logrus.Info(i, ":", j.Keys())
		// }
	}
	db.Close()
	return tables
}
func GenerateTables(dbsetting Model_DBSetting, tables []string) (err error) {
	db, err := ConnectDB(dbsetting)
	if err != nil {
		return
	}

	// db, err := ConnectDB(dbsetting)
	// if err == nil {
	// 	tables = GetAllTables(db, dbsetting.Db)
	// 	GetAllTableAndFields(db, dbsetting.Db)
	// 	// for i, j := range mapfields {
	// 	// 	logrus.Info(i, ":", j.Keys())
	// 	// }
	// }
	/*index data
	[{
		indexname: index1,
		indextype: PRI/INDEX/Unique
		columns: [col1,col2,col3]

	},
	{}]
	*/
	if RunMode != "init" {
		logrus.Info("Skip generate table schemes")
		return
	}
	// tables1 := []string{"acc_payment"}
	for _, tablename := range tables {
		//drop local table if exists
		dropsql := "DROP TABLE IF EXISTS " + tablename
		_, errdrop := localdb.Exec(dropsql)
		if errdrop != nil {
			logrus.Fatal(errdrop)
		}
		var tablesetting []Model_FieldSetting

		sql := "DESCRIBE " + tablename
		// logrus.Info(sql)
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

			sqlcreate := "CREATE TABLE " + tablename + " (`tenant_id` int(11) "

			for _, s := range tablesetting {
				// ignore tenant_id field
				if s.Field == "tenant_id" {
					continue
				}
				// logrus.Info(s.Field, ":", s.Key)
				if s.Key == "PRI" {
					primarykey = "tenant_id," + s.Field
				}

				sqlcreate += ", `" + s.Field + "` " + s.Type
			}

			if primarykey == "" {
				logrus.Warn(tablename + " undefined primarykey")
			}
			// sqlcreate = sqlcreate + ", PRIMARY KEY (" + primarykey + ")) ENGINE=InnoDB"
			sqlcreate = sqlcreate + ") ENGINE=InnoDB"
			logrus.Info("Created table ", tablename)
			res, errcreate := localdb.Query(sqlcreate)
			if errcreate != nil {
				logrus.Fatal(errcreate)
			}
			res.Close()
		} else {
			logrus.Fatal(err2)
		}

	}
	db.Close()
	//logrus.Fatal("Generate tables end")
	return

}

// func CreateFolderIfNotExists(path string) {
// 	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
// 		err := os.Mkdir(path, os.ModePerm)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	}

// }
