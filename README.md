# sqlitemaint

SQLiteMaint package is a specific SQLite database maintenance algorithm.  It works well for simple database project schemas as long as changes are combined into a one file per DB version.

The database maintenance is based on `user_version` SQLite pragma.  When database is 1st created its `user_version` is set to 1.  1st upgrade brings `user_version` to 2.  Next upgrade brings it to 3 and so on.

Versioning conditions:

* `user_version` always starts with 1 when DB is created.
* `user_version` does not skip numbers.

Database creation can be thought of as upgrade from version 0 to version 1.  This makes DB creation and update procedures identical from algorithm point of view.

Each upgrade to the next version number is consolidated in a single upgrade SQL script.  Thus algorithm simply executes the script to bring the database to the next version number.

The following constraints applied to all SQL scripts:

* scripts must be located in the same directory;
* scripts must follow a file name pattern based on the DB version number the script upgrades to.   For example, if script upgrades to `user_version` 3, then script file name is `0003.sql`
* script must upgrade from immediately previous version only.   For example, script `0003.sql` only upgrades from DB version 2.  The script `0003.sql` is not expected to update from DB version 1 or 2.
* the library runs script in a transaction that it creates and controls.

The upgrade procedure thus consists of looping through specified directory with index starting at 0 and sequentially incrementing.  We derive the name of the file from the index.  For example, for index 3 the file name is `0003.sql`.  If file does not exist, it is assumed that DB is fully updated.  If file exists, then script is loaded and executed.  If script generates an error the upgrade process is halted.

If script returns an error, then `user_version` pragma is not updated and transaction is rolledback.  This allows to rerun the upgrade procedure.

## API

API is a single public function:

    func UpgradeSQLite(dbFile, sqlDir string, backup bool) (version int, err error)

`dbFile` is a name of SQLite database file to create or update;
`sqlDir` is a directory that contains SQL scripts to create or update database.
`backup` enables backup of dbFile regardless of weither or not upgrade will occur.  Backed up file will be place into the same directory as `dbFile` and its name will be prefixed with `Copy-of-` prefix.  Algorithm copies file content prior to connection to `dbFile`.  This results in backup on every startup.  (That's something that might be adjusted later).

## Usage

    backup := true
    sqlitemaint.UpgradeSQLite('my.db', 'my', backup)

See [SQLiteMaintainer application](https://github.com/Kulak/sqlitemaintainer) for console application that uses this package.

## License

MIT
