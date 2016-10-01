package sqlitemaint

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	// SQLite driver is required
	_ "github.com/mattn/go-sqlite3"
)

// UpgradeSQLite executes upgrade procedure on SQLite db_file database file
// given sql_dir directory with SQLite upgrade scripts.
func UpgradeSQLite(dbFile, sqlDir string, backup bool) (version int, err error) {
	log.Printf("DB file: %v, SQL files directory: %v\n", dbFile, sqlDir)

	version = -1

	if backup {
		dir, file := path.Split(dbFile)
		backupFile := path.Join(dir, "Copy-of-"+file)
		err = doBackup(dbFile, backupFile)
		if err != nil {
			return
		}
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return
	}
	defer db.Close()

	query := "pragma user_version"
	var userVersion int
	err = db.QueryRow(query).Scan(&userVersion)
	if err != nil {
		err = fmt.Errorf("%q: %s", err, query)
		return
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

func doBackup(src string, dst string) (err error) {
	// Read all content of src to data
	var data []byte
	data, err = ioutil.ReadFile(src)
	if err != nil {
		err = fmt.Errorf("DB backup failed to read DB file %s.  Error: %s", src, err)
		return
	}
	// Write data to dst
	err = ioutil.WriteFile(dst, data, 0644)
	if err != nil {
		err = fmt.Errorf("DB backup failed to write backup file %s.  Error: %s", dst, err)
		return
	}
	return
}
