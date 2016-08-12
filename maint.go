package sqlitemaint

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Executes upgrade procedure on SQLite db_file database file
// given sql_dir directory with SQLite upgrade scripts.
func UpgradeSQLite(db_file, sql_dir string) (version int, err error) {
	log.Printf("DB file: %v, SQL files directory: %v\n", db_file, sql_dir)

	db, err := sql.Open("sqlite3", db_file)
	if err != nil {
		return -1, err
	}
	defer db.Close()

	query := "pragma user_version"
	var user_version int
	err = db.QueryRow(query).Scan(&user_version)
	if err != nil {
		return -1, fmt.Errorf("%q: %s", err, query)
	} else {
		log.Printf("Loaded version: %v\n", user_version)
	}

	user_version++
	update_sql_file := fmt.Sprintf("%v/%04d.sql", sql_dir, user_version)
	_, err = os.Stat(update_sql_file)
	log.Printf("Checking if version file %v exists.\n", update_sql_file)
	// loop while file exists
	for !os.IsNotExist(err) {

		// SQL upgrade file exists

		log.Printf("Processing %v file...\n", update_sql_file)
		var content []byte
		content, err = ioutil.ReadFile(update_sql_file)
		if err != nil {
			return user_version - 1, fmt.Errorf("Failed to read %v file content. Error: %v", update_sql_file, err)
		}
		query = string(content)

		_, err = db.Exec(query)
		if err != nil {
			return user_version - 1, fmt.Errorf("Failed to execute %v file content. Error: %v", update_sql_file, err)
		}

		// pragma values can not be parametrized (in python), so we use sring format and enforce integer value
		query = fmt.Sprintf("pragma user_version = %v", user_version)
		_, err = db.Exec(query)
		if err != nil {
			return user_version - 1, fmt.Errorf("Failed to update DB version to %v after %v file was executed. Error: %v\n",
				user_version, update_sql_file, err)
		} else {
			log.Printf("Updated to version %v.\n", user_version)
		}

		user_version++
		update_sql_file = fmt.Sprintf("%v/%04d.sql", sql_dir, user_version)
		_, err = os.Stat(update_sql_file)
	}
	return user_version - 1, nil
}
