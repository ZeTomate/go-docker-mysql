package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

/**
/*	Main function.
/*	Call all the other functions.
/*		- check the input value
/*		- go throught the input path and retrieve datas
/*		- start the db's call process
**/
func main() {
	var inputName string
	var datas map[string]string = make(map[string]string)

	fmt.Println("start analysing the given parameter...")

	flag.Parse()
	var name = flag.Arg(0)

	// Confirm the input given.
	if name == "" {
		fmt.Println("no file given ?")
		return
	} else {
		inputName = name
		recursiveAnalyse(inputName, datas)
	}

	commandReader := bufio.NewReader(os.Stdin)
	fmt.Print("csv has been parsed... do you want to check values ? (y/N)")
	text, _ := commandReader.ReadString('\n')
	if text == "y\n" {
		fmt.Println(datas)
	}

	dbTransaction(datas)

	fmt.Println("DONE, thx for waiting.")
}


/**
/*	Analyse the given object's path.
/*	If it's a folder, go inside and repeat the process for each child
/*	If it's a file, check if its end bu '.csv', and parse it in this case.
/*	Return the data completed with each csv informations.
**/
func recursiveAnalyse(filePath string, datas map[string]string) map[string]string {
	var file *os.File
	var err error
	var stat os.FileInfo

	// Open object path
	if file, err = os.Open(filePath); err != nil {
		fmt.Println("Can't open the given path...")
		log.Fatal(err)
	}

	// Check object type
	stat, _ = file.Stat()
	if stat.IsDir() {
		files, _ := ioutil.ReadDir(filePath)
		for _, f := range files {
			recursiveAnalyse(filePath+"/"+f.Name(), datas)
		}
	} else if strings.Contains(filePath, ".csv") {
		b, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Print(err)
		}
		var csvDatas = string(b)
		datas = readCsv(csvDatas, datas)
	}
	return datas
}


/**
/*	Parse and extract the correct informations from the CVS.
/*	The CSV should at least have the two fields ['timestamp', 'value'].
/*	Other column will not be taken into consideration.
**/
func readCsv(csvDatas string, datas map[string]string) map[string]string {
	// Parse the given file.
	reader := csv.NewReader(strings.NewReader(csvDatas))
	records, err := reader.ReadAll()
	if err == io.EOF {
	}
	if err != nil {
		log.Fatal(err)
	}
	var tsColumn = indexOf(records[0], "timestamp")
	var vColumn = indexOf(records[0], "value")

	// Store data in a Map to access them easily.
	for _, record := range records[1:] {
		ts := record[tsColumn]
		val := record[vColumn]
		datas[ts] = val
	}
	return datas
}


/**
/*	Return the index of a specific name.
/*	If the value is not found, the function return -1 witch will provoc an
/*	error when trying the access the value in the row's list.
**/
func indexOf(mainArray []string, searchValue string) int {
	for k, v := range mainArray {
		if v == searchValue {
			return k
		}
	}
	return -1
}


/**
/*	Allow to change the base settings if needed.
/*	If the docker image used is the same as me, no changes should be required.
**/
func dbSettings() string {
	commandReader := bufio.NewReader(os.Stdin)

	fmt.Print("Do you want to customise your db settings ? (y/N)")
	text, _ := commandReader.ReadString('\n')

	if text == "y\n" {
		fmt.Print("Please enter the DB name: (QOSenergy)")
		text, _ = commandReader.ReadString('\n')
		db_name := text[:len(text)-1]

		fmt.Print("Please enter the user: (user)")
		text, _ = commandReader.ReadString('\n')
		user := text[:len(text)-1]

		fmt.Print("Please enter the password: (already file)")
		text, _ = commandReader.ReadString('\n')
		psw := text[:len(text)-1]

		return fmt.Sprintf("%s:%s@/%s", user, psw, db_name)
	} else {
		return fmt.Sprintf("%s:%s@/%s", "user", "qos", "QOSenergy")
	}
}


/**
/*	Confirm the Table format.
/*	It doesn't need to be only composed of the two field 'timestamp' and 'value', but those should be there and of the right type.
/*	Return true is the table hit the requirements, false otherwise.
**/
func isTableExistingAndCorrect(db *sql.DB) bool {
	var count int

	columnsNameRows, _ := db.Query("SELECT column_name, column_type FROM information_schema.columns WHERE table_name='Values';")

	for columnsNameRows.Next() {
		var (
			column_name string
			column_type string
		)
		if err := columnsNameRows.Scan(&column_name, &column_type); err != nil {
			log.Fatal(err)
			panic(err.Error())
		}
		if (column_name == "timestamp" && column_type == "int(11)") || (column_name == "value" && column_type == "float") {
			count++
		}
	}
	if count == 2 {
		return true
	}
	return false
}


/**
/*	Manage the full interaction with the database, from the settings to the insert.
**/
func dbTransaction(datas map[string]string) {
	// Confirme/Update the database settings.
	dataSourceName := dbSettings()

	// Connect to the db
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	// Create the correct table if not created before hand.
	_, creationErr := db.Query("CREATE TABLE IF NOT EXISTS `Values` (`timestamp` INT NOT NULL, `value` FLOAT NOT NULL);")
	if creationErr != nil {
		panic(creationErr.Error())
	}

	// Confirm the table's existence and its format.
	if !isTableExistingAndCorrect(db) {
		fmt.Println("Bad table: insert not executed.")
		return
	}

	// Prepare the insert request.
	sqlStr := "INSERT INTO `Values` (`timestamp`, `value`) VALUES "
	vals := []interface{}{}

	for k, v := range datas {
		sqlStr += "(?, ?), "
		vals = append(vals, k, v)
	}
	sqlStr = sqlStr[0 : len(sqlStr)-2]

	stmt, _ := db.Prepare(sqlStr)

	// insert the actuals values.
	_, insertErr := stmt.Exec(vals...)
	if insertErr != nil {
		fmt.Println("NOT IMPORTED !")
		panic(insertErr.Error())
	}
}
