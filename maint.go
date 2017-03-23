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
		if pathExists(dbFile) {
			dir, file := path.Split(dbFile)
			backupFile := path.Join(dir, "Copy-of-"+file)
			err = doBackup(dbFile, backupFile)
			if err != nil {
				return
			}
		}
	}

	dbc, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return
	}
	defer dbc.Close()

	query := "pragma user_version"
	var userVersion int
	err = dbc.QueryRow(query).Scan(&userVersion)
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

		// create transaction
		var tx *sql.Tx
		tx, err = dbc.Begin()

		log.Printf("Processing %v file...\n", fileName)
		var content []byte
		content, err = ioutil.ReadFile(fileName)
		if err != nil {
			err = fmt.Errorf("Failed to read %v file content. Error: %v", fileName, err)
			goto ROLLBACK
		}
		query = string(content)

		// execute update script
		_, err = tx.Exec(query)
		if err != nil {
			err = fmt.Errorf("Failed to execute %v file content. Error: %v", fileName, err)
			goto ROLLBACK
		}

		// pragma values can not be parametrized (in python), so we use sring format and enforce integer value
		query = fmt.Sprintf("pragma user_version = %v", userVersion)
		_, err = tx.Exec(query)
		if err != nil {
			err = fmt.Errorf("Failed to update DB version to %v after %v file was executed. Error: %v\n", userVersion, fileName, err)
			goto ROLLBACK
		}
		log.Printf("Updated to version %v.\n", userVersion)

	ROLLBACK:
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				fmt.Printf("Failed to rollback transaction: %s", err2)
			}
			return userVersion - 1, err
		}
		err = tx.Commit()
		if err != nil {
			fmt.Printf("Error on Commit: %s", err)
			err = tx.Rollback()
			if err != nil {
				fmt.Printf("Error on rollback: %s", err)
				return userVersion - 1, err
			}
		}

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

// PathExists returns true if file or directory exists.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}
