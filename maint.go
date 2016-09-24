package sqlitemaint

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	// SQLite driver is required
	_ "github.com/mattn/go-sqlite3"
)

// UpgradeSQLite executes upgrade procedure on SQLite db_file database file
// given sql_dir directory with SQLite upgrade scripts.
func UpgradeSQLite(dbFile, sqlDir string) (version int, err error) {
	log.Printf("DB file: %v, SQL files directory: %v\n", dbFile, sqlDir)

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return -1, err
	}
	defer db.Close()

	query := "pragma user_version"
	var userVersion int
	err = db.QueryRow(query).Scan(&userVersion)
	if err != nil {
		return -1, fmt.Errorf("%q: %s", err, query)
	}
	log.Printf("Loaded version: %v\n", userVersion)

	userVersion++
	fileName := fmt.Sprintf("%v/%04d.sql", sqlDir, userVersion)
	_, err = os.Stat(fileName)
	log.Printf("Checking if version file %v exists.\n", fileName)
	// loop while file exists
	for !os.IsNotExist(err) {

		// SQL upgrade file exists

		log.Printf("Processing %v file...\n", fileName)
		var content []byte
		content, err = ioutil.ReadFile(fileName)
		if err != nil {
			return userVersion - 1, fmt.Errorf("Failed to read %v file content. Error: %v", fileName, err)
		}
		query = string(content)

		_, err = db.Exec(query)
		if err != nil {
			return userVersion - 1, fmt.Errorf("Failed to execute %v file content. Error: %v", fileName, err)
		}

		// pragma values can not be parametrized (in python), so we use sring format and enforce integer value
		query = fmt.Sprintf("pragma user_version = %v", userVersion)
		_, err = db.Exec(query)
		if err != nil {
			return userVersion - 1, fmt.Errorf("Failed to update DB version to %v after %v file was executed. Error: %v\n",
				userVersion, fileName, err)
		}
		log.Printf("Updated to version %v.\n", userVersion)

		userVersion++
		fileName = fmt.Sprintf("%v/%04d.sql", sqlDir, userVersion)
		_, err = os.Stat(fileName)
	}
	return userVersion - 1, nil
}
