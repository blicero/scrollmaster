// /home/krylon/go/src/github.com/blicero/scrollmaster/database/database.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 08. 2024 by Benjamin Walkenhorst
// (c) 2024 Benjamin Walkenhorst
// Time-stamp: <2024-09-13 18:09:09 krylon>

package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/blicero/krylib"
	"github.com/blicero/scrollmaster/common"
	"github.com/blicero/scrollmaster/database/query"
	"github.com/blicero/scrollmaster/logdomain"
	"github.com/blicero/scrollmaster/model"
	_ "github.com/mattn/go-sqlite3" // Import the database driver
)

var (
	openLock sync.Mutex
	idCnt    int64
)

// ErrTxInProgress indicates that an attempt to initiate a transaction failed
// because there is already one in progress.
var ErrTxInProgress = errors.New("A Transaction is already in progress")

// ErrNoTxInProgress indicates that an attempt was made to finish a
// transaction when none was active.
var ErrNoTxInProgress = errors.New("There is no transaction in progress")

// ErrEmptyUpdate indicates that an update operation would not change any
// values.
var ErrEmptyUpdate = errors.New("Update operation does not change any values")

// ErrInvalidValue indicates that one or more parameters passed to a method
// had values that are invalid for that operation.
var ErrInvalidValue = errors.New("Invalid value for parameter")

// ErrObjectNotFound indicates that an Object was not found in the database.
var ErrObjectNotFound = errors.New("object was not found in database")

// ErrInvalidSavepoint is returned when a user of the Database uses an unkown
// (or expired) savepoint name.
var ErrInvalidSavepoint = errors.New("that save point does not exist")

// If a query returns an error and the error text is matched by this regex, we
// consider the error as transient and try again after a short delay.
var retryPat = regexp.MustCompile("(?i)database is (?:locked|busy)")

// worthARetry returns true if an error returned from the database
// is matched by the retryPat regex.
func worthARetry(e error) bool {
	return retryPat.MatchString(e.Error())
} // func worthARetry(e error) bool

// retryDelay is the amount of time we wait before we repeat a database
// operation that failed due to a transient error.
const retryDelay = 25 * time.Millisecond

func waitForRetry() {
	time.Sleep(retryDelay)
} // func waitForRetry()

// Database wraps a database connection and associated state.
type Database struct {
	id            int64
	db            *sql.DB
	tx            *sql.Tx
	log           *log.Logger
	path          string
	spNameCounter int
	spNameCache   map[string]string
	queries       map[query.ID]*sql.Stmt
}

// Open opens a Database. If the database specified by the path does not exist,
// yet, it is created and initialized.
func Open(path string) (*Database, error) {
	var (
		err      error
		dbExists bool
		db       = &Database{
			path:          path,
			spNameCounter: 1,
			spNameCache:   make(map[string]string),
			queries:       make(map[query.ID]*sql.Stmt),
		}
	)

	openLock.Lock()
	defer openLock.Unlock()
	idCnt++
	db.id = idCnt

	if db.log, err = common.GetLogger(logdomain.Database); err != nil {
		return nil, err
	} else if common.Debug {
		db.log.Printf("[DEBUG] Open database %s\n", path)
	}

	var connstring = fmt.Sprintf("%s?_locking=NORMAL&_journal=WAL&_fk=true&recursive_triggers=true",
		path)

	if dbExists, err = krylib.Fexists(path); err != nil {
		db.log.Printf("[ERROR] Failed to check if %s already exists: %s\n",
			path,
			err.Error())
		return nil, err
	} else if db.db, err = sql.Open("sqlite3", connstring); err != nil {
		db.log.Printf("[ERROR] Failed to open %s: %s\n",
			path,
			err.Error())
		return nil, err
	}

	if !dbExists {
		if err = db.initialize(); err != nil {
			var e2 error
			if e2 = db.db.Close(); e2 != nil {
				db.log.Printf("[CRITICAL] Failed to close database: %s\n",
					e2.Error())
				return nil, e2
			} else if e2 = os.Remove(path); e2 != nil {
				db.log.Printf("[CRITICAL] Failed to remove database file %s: %s\n",
					db.path,
					e2.Error())
			}
			return nil, err
		}
		db.log.Printf("[INFO] Database at %s has been initialized\n",
			path)
	}

	return db, nil
} // func Open(path string) (*Database, error)

func (db *Database) initialize() error {
	var err error
	var tx *sql.Tx

	if common.Debug {
		db.log.Printf("[DEBUG] Initialize fresh database at %s\n",
			db.path)
	}

	if tx, err = db.db.Begin(); err != nil {
		db.log.Printf("[ERROR] Cannot begin transaction: %s\n",
			err.Error())
		return err
	}

	for _, q := range qInit {
		db.log.Printf("[TRACE] Execute init query:\n%s\n",
			q)
		if _, err = tx.Exec(q); err != nil {
			db.log.Printf("[ERROR] Cannot execute init query: %s\n%s\n",
				err.Error(),
				q)
			if rbErr := tx.Rollback(); rbErr != nil {
				db.log.Printf("[CANTHAPPEN] Cannot rollback transaction: %s\n",
					rbErr.Error())
				return rbErr
			}
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		db.log.Printf("[CANTHAPPEN] Failed to commit init transaction: %s\n",
			err.Error())
		return err
	}

	return nil
} // func (db *Database) initialize() error

// Close closes the database.
// If there is a pending transaction, it is rolled back.
func (db *Database) Close() error {
	// I wonder if would make more snese to panic() if something goes wrong

	var err error

	if db.tx != nil {
		if err = db.tx.Rollback(); err != nil {
			db.log.Printf("[CRITICAL] Cannot roll back pending transaction: %s\n",
				err.Error())
			return err
		}
		db.tx = nil
	}

	for key, stmt := range db.queries {
		if err = stmt.Close(); err != nil {
			db.log.Printf("[CRITICAL] Cannot close statement handle %s: %s\n",
				key,
				err.Error())
			return err
		}
		delete(db.queries, key)
	}

	if err = db.db.Close(); err != nil {
		db.log.Printf("[CRITICAL] Cannot close database: %s\n",
			err.Error())
	}

	db.db = nil
	return nil
} // func (db *Database) Close() error

func (db *Database) getQuery(id query.ID) (*sql.Stmt, error) {
	var (
		stmt  *sql.Stmt
		found bool
		err   error
	)

	if stmt, found = db.queries[id]; found {
		return stmt, nil
	} else if _, found = qdb[id]; !found {
		return nil, fmt.Errorf("Unknown Query %d",
			id)
	}

	db.log.Printf("[TRACE] Prepare query %s\n", id)

PREPARE_QUERY:
	if stmt, err = db.db.Prepare(qdb[id]); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto PREPARE_QUERY
		}

		db.log.Printf("[ERROR] Cannot parse query %s: %s\n%s\n",
			id,
			err.Error(),
			qdb[id])
		return nil, err
	}

	db.queries[id] = stmt
	return stmt, nil
} // func (db *Database) getQuery(query.ID) (*sql.Stmt, error)

func (db *Database) resetSPNamespace() {
	db.spNameCounter = 1
	db.spNameCache = make(map[string]string)
} // func (db *Database) resetSPNamespace()

func (db *Database) generateSPName(name string) string {
	var spname = fmt.Sprintf("Savepoint%05d",
		db.spNameCounter)

	db.spNameCache[name] = spname
	db.spNameCounter++
	return spname
} // func (db *Database) generateSPName() string

// PerformMaintenance performs some maintenance operations on the database.
// It cannot be called while a transaction is in progress and will block
// pretty much all access to the database while it is running.
func (db *Database) PerformMaintenance() error {
	var mQueries = []string{
		"PRAGMA wal_checkpoint(TRUNCATE)",
		"VACUUM",
		"REINDEX",
		"ANALYZE",
	}
	var err error

	if db.tx != nil {
		return ErrTxInProgress
	}

	for _, q := range mQueries {
		if _, err = db.db.Exec(q); err != nil {
			db.log.Printf("[ERROR] Failed to execute %s: %s\n",
				q,
				err.Error())
		}
	}

	return nil
} // func (db *Database) PerformMaintenance() error

// Begin begins an explicit database transaction.
// Only one transaction can be in progress at once, attempting to start one,
// while another transaction is already in progress will yield ErrTxInProgress.
func (db *Database) Begin() error {
	var err error

	db.log.Printf("[DEBUG] Database#%d Begin Transaction\n",
		db.id)

	if db.tx != nil {
		return ErrTxInProgress
	}

BEGIN_TX:
	for db.tx == nil {
		if db.tx, err = db.db.Begin(); err != nil {
			if worthARetry(err) {
				waitForRetry()
				continue BEGIN_TX
			} else {
				db.log.Printf("[ERROR] Failed to start transaction: %s\n",
					err.Error())
				return err
			}
		}
	}

	db.resetSPNamespace()

	return nil
} // func (db *Database) Begin() error

// SavepointCreate creates a savepoint with the given name.
//
// Savepoints only make sense within a running transaction, and just like
// with explicit transactions, managing them is the responsibility of the
// user of the Database.
//
// Creating a savepoint without a surrounding transaction is not allowed,
// even though SQLite allows it.
//
// For details on how Savepoints work, check the excellent SQLite
// documentation, but here's a quick guide:
//
// Savepoints are kind-of-like transactions within a transaction: One
// can create a savepoint, make some changes to the database, and roll
// back to that savepoint, discarding all changes made between
// creating the savepoint and rolling back to it. Savepoints can be
// quite useful, but there are a few things to keep in mind:
//
//   - Savepoints exist within a transaction. When the surrounding transaction
//     is finished, all savepoints created within that transaction cease to exist,
//     no matter if the transaction is commited or rolled back.
//
//   - When the database is recovered after being interrupted during a
//     transaction, e.g. by a power outage, the entire transaction is rolled back,
//     including all savepoints that might exist.
//
//   - When a savepoint is released, nothing changes in the state of the
//     surrounding transaction. That means rolling back the surrounding
//     transaction rolls back the entire transaction, regardless of any
//     savepoints within.
//
//   - Savepoints do not nest. Releasing a savepoint releases it and *all*
//     existing savepoints that have been created before it. Rolling back to a
//     savepoint removes that savepoint and all savepoints created after it.
func (db *Database) SavepointCreate(name string) error {
	var err error

	db.log.Printf("[DEBUG] SavepointCreate(%s)\n",
		name)

	if db.tx == nil {
		return ErrNoTxInProgress
	}

SAVEPOINT:
	// It appears that the SAVEPOINT statement does not support placeholders.
	// But I do want to used named savepoints.
	// And I do want to use the given name so that no SQL injection
	// becomes possible.
	// It would be nice if the database package or at least the SQLite
	// driver offered a way to escape the string properly.
	// One possible solution would be to use names generated by the
	// Database instead of user-defined names.
	//
	// But then I need a way to use the Database-generated name
	// in rolling back and releasing the savepoint.
	// I *could* use the names strictly inside the Database, store them in
	// a map or something and hand out a key to that name to the user.
	// Since savepoint only exist within one transaction, I could even
	// re-use names from one transaction to the next.
	//
	// Ha! I could accept arbitrary names from the user, generate a
	// clean name, and store these in a map. That way the user can
	// still choose names that are outwardly visible, but they do
	// not touch the Database itself.
	//
	//if _, err = db.tx.Exec("SAVEPOINT ?", name); err != nil {
	// if _, err = db.tx.Exec("SAVEPOINT " + name); err != nil {
	// 	if worthARetry(err) {
	// 		waitForRetry()
	// 		goto SAVEPOINT
	// 	}

	// 	db.log.Printf("[ERROR] Failed to create savepoint %s: %s\n",
	// 		name,
	// 		err.Error())
	// }

	var internalName = db.generateSPName(name)

	var spQuery = "SAVEPOINT " + internalName

	if _, err = db.tx.Exec(spQuery); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto SAVEPOINT
		}

		db.log.Printf("[ERROR] Failed to create savepoint %s: %s\n",
			name,
			err.Error())
	}

	return err
} // func (db *Database) SavepointCreate(name string) error

// SavepointRelease releases the Savepoint with the given name, and all
// Savepoints created before the one being release.
func (db *Database) SavepointRelease(name string) error {
	var (
		err                   error
		internalName, spQuery string
		validName             bool
	)

	db.log.Printf("[DEBUG] SavepointRelease(%s)\n",
		name)

	if db.tx != nil {
		return ErrNoTxInProgress
	}

	if internalName, validName = db.spNameCache[name]; !validName {
		db.log.Printf("[ERROR] Attempt to release unknown Savepoint %q\n",
			name)
		return ErrInvalidSavepoint
	}

	db.log.Printf("[DEBUG] Release Savepoint %q (%q)",
		name,
		db.spNameCache[name])

	spQuery = "RELEASE SAVEPOINT " + internalName

SAVEPOINT:
	if _, err = db.tx.Exec(spQuery); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto SAVEPOINT
		}

		db.log.Printf("[ERROR] Failed to release savepoint %s: %s\n",
			name,
			err.Error())
	} else {
		delete(db.spNameCache, internalName)
	}

	return err
} // func (db *Database) SavepointRelease(name string) error

// SavepointRollback rolls back the running transaction to the given savepoint.
func (db *Database) SavepointRollback(name string) error {
	var (
		err                   error
		internalName, spQuery string
		validName             bool
	)

	db.log.Printf("[DEBUG] SavepointRollback(%s)\n",
		name)

	if db.tx != nil {
		return ErrNoTxInProgress
	}

	if internalName, validName = db.spNameCache[name]; !validName {
		return ErrInvalidSavepoint
	}

	spQuery = "ROLLBACK TO SAVEPOINT " + internalName

SAVEPOINT:
	if _, err = db.tx.Exec(spQuery); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto SAVEPOINT
		}

		db.log.Printf("[ERROR] Failed to create savepoint %s: %s\n",
			name,
			err.Error())
	}

	delete(db.spNameCache, name)
	return err
} // func (db *Database) SavepointRollback(name string) error

// Rollback terminates a pending transaction, undoing any changes to the
// database made during that transaction.
// If no transaction is active, it returns ErrNoTxInProgress
func (db *Database) Rollback() error {
	var err error

	db.log.Printf("[DEBUG] Database#%d Roll back Transaction\n",
		db.id)

	if db.tx == nil {
		return ErrNoTxInProgress
	} else if err = db.tx.Rollback(); err != nil {
		return fmt.Errorf("Cannot roll back database transaction: %s",
			err.Error())
	}

	db.tx = nil
	db.resetSPNamespace()

	return nil
} // func (db *Database) Rollback() error

// Commit ends the active transaction, making any changes made during that
// transaction permanent and visible to other connections.
// If no transaction is active, it returns ErrNoTxInProgress
func (db *Database) Commit() error {
	var err error

	db.log.Printf("[DEBUG] Database#%d Commit Transaction\n",
		db.id)

	if db.tx == nil {
		return ErrNoTxInProgress
	} else if err = db.tx.Commit(); err != nil {
		return fmt.Errorf("Cannot commit transaction: %s",
			err.Error())
	}

	db.resetSPNamespace()
	db.tx = nil
	return nil
} // func (db *Database) Commit() error

// HostAdd adds a new Host to the database.
func (db *Database) HostAdd(h *model.Host) error {
	const qid query.ID = query.HostAdd
	var (
		err    error
		msg    string
		stmt   *sql.Stmt
		tx     *sql.Tx
		status bool
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid.String(),
			err.Error())
		return err
	} else if db.tx != nil {
		tx = db.tx
	} else {
		db.log.Printf("[INFO] Start ad-hoc transaction for adding Host %s\n",
			h.Name)
	BEGIN_AD_HOC:
		if tx, err = db.db.Begin(); err != nil {
			if worthARetry(err) {
				waitForRetry()
				goto BEGIN_AD_HOC
			} else {
				msg = fmt.Sprintf("Error starting transaction: %s\n",
					err.Error())
				db.log.Printf("[ERROR] %s\n", msg)
				return errors.New(msg)
			}

		} else {
			defer func() {
				var err2 error
				if status {
					if err2 = tx.Commit(); err2 != nil {
						db.log.Printf("[ERROR] Failed to commit ad-hoc transaction: %s\n",
							err2.Error())
					}
				} else if err2 = tx.Rollback(); err2 != nil {
					db.log.Printf("[ERROR] Rollback of ad-hoc transaction failed: %s\n",
						err2.Error())
				}
			}()
		}
	}

	stmt = tx.Stmt(stmt)
	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(h.Name, h.LastSeen.Unix()); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("Cannot add Host %s to database: %s",
				h.Name,
				err.Error())
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	} else {
		var id int64

		defer rows.Close()

		if !rows.Next() {
			// CANTHAPPEN
			db.log.Printf("[ERROR] Query %s did not return a value\n",
				qid)
			return fmt.Errorf("Query %s did not return a value", qid)
		} else if err = rows.Scan(&id); err != nil {
			msg = fmt.Sprintf("Failed to get ID for newly added host %s: %s",
				h.Name,
				err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return errors.New(msg)
		}

		h.ID = id
		status = true
		return nil
	}
} // func (db *Database) HostAdd(h *model.Host) error

// HostGetByName fetches a Host by its name.
func (db *Database) HostGetByName(name string) (*model.Host, error) {
	const qid query.ID = query.HostGetByName
	var (
		err  error
		msg  string
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(name); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	if rows.Next() {
		var (
			timestamp int64
			h         = &model.Host{Name: name}
		)

		if err = rows.Scan(&h.ID, &timestamp); err != nil {
			msg = fmt.Sprintf("Error scanning row for Host %s: %s",
				name,
				err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return nil, errors.New(msg)
		}

		h.LastSeen = time.Unix(timestamp, 0)
		return h, nil
	}

	db.log.Printf("[INFO] Host %s was not found in database\n", name)
	return nil, nil
} // func (db *Database) HostGetByName(name string) (*model.Host, error)

// HostGetByID fetches a Host by its ID.
func (db *Database) HostGetByID(id int64) (*model.Host, error) {
	const qid query.ID = query.HostGetByID
	var (
		err  error
		msg  string
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(id); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	if rows.Next() {
		var (
			timestamp int64
			h         = &model.Host{ID: id}
		)

		if err = rows.Scan(&h.Name, &timestamp); err != nil {
			msg = fmt.Sprintf("Error scanning row for Host %d: %s",
				id,
				err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return nil, errors.New(msg)
		}

		h.LastSeen = time.Unix(timestamp, 0)
		return h, nil
	}

	db.log.Printf("[INFO] Host %d was not found in database\n", id)
	return nil, nil
} // func (db *Database) HostGetByID(id int64) (*model.Host, error)

// HostGetAll fetches all Hosts
func (db *Database) HostGetAll() ([]model.Host, error) {
	const qid query.ID = query.HostGetAll
	var (
		err  error
		msg  string
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var hosts = make([]model.Host, 0)

	for rows.Next() {
		var (
			h         model.Host
			timestamp int64
		)

		if err = rows.Scan(&h.ID, &h.Name, &timestamp); err != nil {
			msg = fmt.Sprintf("Failed to scan row: %s", err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return nil, errors.New(msg)
		}

		h.LastSeen = time.Unix(timestamp, 0)
		hosts = append(hosts, h)
	}

	return hosts, nil
} // func (db *Database) HostGetAll() ([]model.Host, error)

// HostUpdateLastSeen updates the timestamp a Host was last seen by the Server.
func (db *Database) HostUpdateLastSeen(h *model.Host, timestamp time.Time) error {
	const qid query.ID = query.HostUpdateLastSeen
	var (
		err    error
		msg    string
		stmt   *sql.Stmt
		tx     *sql.Tx
		status bool
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid.String(),
			err.Error())
		return err
	} else if db.tx != nil {
		tx = db.tx
	} else {
		db.log.Printf("[INFO] Start ad-hoc transaction for adding Host %s\n",
			h.Name)
	BEGIN_AD_HOC:
		if tx, err = db.db.Begin(); err != nil {
			if worthARetry(err) {
				waitForRetry()
				goto BEGIN_AD_HOC
			} else {
				msg = fmt.Sprintf("Error starting transaction: %s\n",
					err.Error())
				db.log.Printf("[ERROR] %s\n", msg)
				return errors.New(msg)
			}

		} else {
			defer func() {
				var err2 error
				if status {
					if err2 = tx.Commit(); err2 != nil {
						db.log.Printf("[ERROR] Failed to commit ad-hoc transaction: %s\n",
							err2.Error())
					}
				} else if err2 = tx.Rollback(); err2 != nil {
					db.log.Printf("[ERROR] Rollback of ad-hoc transaction failed: %s\n",
						err2.Error())
				}
			}()
		}
	}

	stmt = tx.Stmt(stmt)

EXEC_QUERY:
	if _, err = stmt.Exec(timestamp.Unix(), h.ID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("Cannot add Host %s to database: %s",
				h.Name,
				err.Error())
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	}

	status = true
	h.LastSeen = timestamp
	return nil
} // func (db *Database) HostUpdateLastSeen(h *model.Host, timestamp time.Time) error

// RecordAdd adds a new Record to the Database.
func (db *Database) RecordAdd(r *model.Record) error {
	const qid query.ID = query.RecordAdd
	var (
		err    error
		msg    string
		stmt   *sql.Stmt
		tx     *sql.Tx
		status bool
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid.String(),
			err.Error())
		return err
	} else if db.tx != nil {
		tx = db.tx
	} else {
		db.log.Println("[INFO] Start ad-hoc transaction for adding Record.")
	BEGIN_AD_HOC:
		if tx, err = db.db.Begin(); err != nil {
			if worthARetry(err) {
				waitForRetry()
				goto BEGIN_AD_HOC
			} else {
				msg = fmt.Sprintf("Error starting transaction: %s\n",
					err.Error())
				db.log.Printf("[ERROR] %s\n", msg)
				return errors.New(msg)
			}

		} else {
			defer func() {
				var err2 error
				if status {
					if err2 = tx.Commit(); err2 != nil {
						db.log.Printf("[ERROR] Failed to commit ad-hoc transaction: %s\n",
							err2.Error())
					}
				} else if err2 = tx.Rollback(); err2 != nil {
					db.log.Printf("[ERROR] Rollback of ad-hoc transaction failed: %s\n",
						err2.Error())
				}
			}()
		}
	}

	stmt = tx.Stmt(stmt)
	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(r.HostID, r.Time.Unix(), r.Source, r.Message, r.Checksum()); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("Cannot add Record to database: %s",
				err.Error())
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	}

	var id int64

	defer rows.Close()

	if !rows.Next() {
		// CANTHAPPEN
		db.log.Printf("[CANTHAPPEN] Query %s did not return a value\n\n%#v\n\n",
			qid,
			r)
		return fmt.Errorf("Query %s did not return a value", qid)
	} else if err = rows.Scan(&id); err != nil {
		msg = fmt.Sprintf("Failed to get ID for newly added Record: %s",
			err.Error())
		db.log.Printf("[ERROR] %s\n", msg)
		return errors.New(msg)
	}

	r.ID = id
	status = true
	return nil
} // func (db *Database) RecordAdd(r *model.Record) error

// RecordGetByHost fetches the <max> most recent records for a given Host.
func (db *Database) RecordGetByHost(h *model.Host, max int64) ([]model.Record, error) {
	const qid query.ID = query.RecordGetByHost
	var (
		err  error
		msg  string
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(h.ID, max); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var records = make([]model.Record, 0)

	for rows.Next() {
		var (
			r         = model.Record{HostID: h.ID}
			timestamp int64
		)

		if err = rows.Scan(&r.ID, &timestamp, &r.Source, &r.Message); err != nil {
			msg = fmt.Sprintf("Failed to scan row: %s", err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return nil, errors.New(msg)
		}

		r.Time = time.Unix(timestamp, 0)
		records = append(records, r)
	}

	//slices.Reverse(records)

	return records, nil
} // func (db *Database) RecordGetByHost(h *model.Host, max int64) ([]model.Record, error)

// RecordGetByPeriod fetches all records for the given period, ordered by their timestamps.
func (db *Database) RecordGetByPeriod(begin, end time.Time) ([]model.Record, error) {
	const qid query.ID = query.RecordGetByPeriod
	var (
		err  error
		msg  string
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(begin.Unix(), end.Unix()); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var records = make([]model.Record, 0)

	for rows.Next() {
		var (
			r         model.Record
			timestamp int64
		)

		if err = rows.Scan(&r.ID, &timestamp, &r.Source, &r.Message); err != nil {
			msg = fmt.Sprintf("Failed to scan row: %s", err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return nil, errors.New(msg)
		}

		r.Time = time.Unix(timestamp, 0)
		records = append(records, r)
	}

	return records, nil
} // func (db *Database) RecordGetByPeriod(begin, end time.Time) ([]model.Record, error)

// RecordGetMostRecent returns the timestamp of the youngest record for a given host.
func (db *Database) RecordGetMostRecent(hostID int64) (time.Time, error) {
	const qid query.ID = query.RecordGetMostRecent
	var (
		err   error
		msg   string
		stmt  *sql.Stmt
		stamp time.Time
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return stamp, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(hostID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return stamp, err
	}

	defer rows.Close() // nolint: errcheck

	if !rows.Next() {
		msg = fmt.Sprintf("No records for Host %d", hostID)
		db.log.Printf("[INFO] %s\n", msg)
		return stamp, errors.New(msg)
	}

	var seconds int64

	if err = rows.Scan(&seconds); err != nil {
		msg = fmt.Sprintf("Failed to extract timestamp (INTEGER) from database: %s",
			err.Error())
		db.log.Printf("[ERROR] %s\n", msg)
		return stamp, errors.New(msg)
	}

	stamp = time.Unix(seconds, 0)
	return stamp, nil
} // func (db *Database) RecordGetMostRecent(hostID int64) (time.Time, error)

// RecordCheckExist returns true if a record with the given checksum exists in the database
func (db *Database) RecordCheckExist(cksum string) (bool, error) {
	const qid query.ID = query.RecordCheckExist
	var (
		err   error
		msg   string
		stmt  *sql.Stmt
		exist bool
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return exist, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(cksum); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return exist, err
	}

	defer rows.Close() // nolint: errcheck

	if !rows.Next() {
		// CANTHAPPEN
		msg = fmt.Sprintf("No records for Checksum %s", cksum)
		db.log.Printf("[CANTHAPPEN] %s\n", msg)
		return false, errors.New(msg)
	}

	var cnt int64

	if err = rows.Scan(&cnt); err != nil {
		msg = fmt.Sprintf("Failed to extract timestamp (INTEGER) from database: %s",
			err.Error())
		db.log.Printf("[ERROR] %s\n", msg)
		return exist, errors.New(msg)
	}

	exist = cnt > 0
	return exist, nil
} // func (db *Database) RecordCheckExist(cksum string) (bool, error)

// RecordGetRecent fetches the <max> most recent Records from the database.
// A negative number returns ALL records, which should probably be avoided.
func (db *Database) RecordGetRecent(max int64) ([]model.Record, error) {
	const qid query.ID = query.RecordGetRecent
	var (
		err  error
		msg  string
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(max); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var records = make([]model.Record, 0)

	for rows.Next() {
		var (
			r         model.Record
			timestamp int64
		)

		if err = rows.Scan(&r.ID, &r.HostID, &timestamp, &r.Source, &r.Message); err != nil {
			msg = fmt.Sprintf("Failed to scan row: %s", err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return nil, errors.New(msg)
		}

		r.Time = time.Unix(timestamp, 0)
		records = append(records, r)
	}

	return records, nil
} // func (db *Database) RecordGetRecent(max int64) ([]model.Record, error)

// RecordGetSources returns a map of all the distinct sources occurring in the database
// and their respective frequencies.
func (db *Database) RecordGetSources() (map[string]int64, error) {
	const qid query.ID = query.RecordGetSources
	var (
		err  error
		msg  string
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var sources = make(map[string]int64)

	for rows.Next() {
		var (
			src string
			cnt int64
		)

		if err = rows.Scan(&src, &cnt); err != nil {
			msg = fmt.Sprintf("Failed to scan row: %s", err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return nil, errors.New(msg)
		}

		sources[src] = cnt
	}

	return sources, nil
} // func (db *Database) RecordGetSources() (map[string]int64, error)

// RecordSearch searches ALL Records in the database according to the query.
func (db *Database) RecordSearch(search *model.SearchQuery, q chan<- model.Record) {
	const qid query.ID = query.RecordGetRecent
	var (
		err  error
		msg  string
		stmt *sql.Stmt
	)

	defer close(q)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(-1); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return
	}

	defer rows.Close() // nolint: errcheck,gosec

	for rows.Next() {
		var (
			r         model.Record
			timestamp int64
		)

		if err = rows.Scan(&r.ID, &r.HostID, &timestamp, &r.Source, &r.Message); err != nil {
			msg = fmt.Sprintf("Failed to scan row: %s", err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return
		}

		r.Time = time.Unix(timestamp, 0)
		if search.Match(&r) {
			q <- r
		}
	}
} // func (db *Database) RecordSearch(search *model.SearchQuery, q chan<- model.Record)

// SearchAdd adds a Search to the database, including both the query and the results.
func (db *Database) SearchAdd(search *model.Search) error {
	const qid query.ID = query.SearchAdd
	var (
		err                  error
		msg                  string
		stmt                 *sql.Stmt
		tx                   *sql.Tx
		bufQuery, bufResults []byte
		status               bool
	)

	if bufQuery, err = json.Marshal(&search.Query); err != nil {
		db.log.Printf("[ERROR] Cannot serialize Query to JSON: %s\n",
			err.Error())
		return err
	} else if bufResults, err = json.Marshal(search.Results); err != nil {
		db.log.Printf("[ERROR] Cannot serialize Results to JSON: %s\n",
			err.Error())
	}

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid.String(),
			err.Error())
		return err
	} else if db.tx != nil {
		tx = db.tx
	} else {
		db.log.Println("[INFO] Start ad-hoc transaction for adding Search.")
	BEGIN_AD_HOC:
		if tx, err = db.db.Begin(); err != nil {
			if worthARetry(err) {
				waitForRetry()
				goto BEGIN_AD_HOC
			} else {
				msg = fmt.Sprintf("Error starting transaction: %s\n",
					err.Error())
				db.log.Printf("[ERROR] %s\n", msg)
				return errors.New(msg)
			}

		} else {
			defer func() {
				var err2 error
				if status {
					if err2 = tx.Commit(); err2 != nil {
						db.log.Printf("[ERROR] Failed to commit ad-hoc transaction: %s\n",
							err2.Error())
					}
				} else if err2 = tx.Rollback(); err2 != nil {
					db.log.Printf("[ERROR] Rollback of ad-hoc transaction failed: %s\n",
						err2.Error())
				}
			}()
		}
	}

	stmt = tx.Stmt(stmt)
	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(
		search.Timestamp.Unix(),
		string(bufQuery),
		string(bufResults),
		len(search.Results)); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("Cannot add Search to database: %s",
				err.Error())
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	}

	var id int64

	defer rows.Close()

	if !rows.Next() {
		// CANTHAPPEN
		err = fmt.Errorf("Query %s did not return a value", qid)
		db.log.Printf("[CANTHAPPEN] %s\n", err.Error())
		return err
	} else if err = rows.Scan(&id); err != nil {
		err = fmt.Errorf("Failed to get ID for newly added Search: %s",
			err.Error())
		db.log.Printf("[ERROR] %s\n", err.Error())
		return err
	}

	search.ID = id
	status = true
	return nil
} // func (db *Database) SearchAdd(search *model.Search) error

// SearchDelete removes a Search from the database.
func (db *Database) SearchDelete(id int64) error {
	const qid query.ID = query.SearchDelete

	var (
		err    error
		msg    string
		stmt   *sql.Stmt
		tx     *sql.Tx
		status bool
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid.String(),
			err.Error())
		return err
	} else if db.tx != nil {
		tx = db.tx
	} else {
		db.log.Printf("[INFO] Start ad-hoc transaction for deleting Search %d.\n", id)
	BEGIN_AD_HOC:
		if tx, err = db.db.Begin(); err != nil {
			if worthARetry(err) {
				waitForRetry()
				goto BEGIN_AD_HOC
			} else {
				msg = fmt.Sprintf("Error starting transaction: %s\n",
					err.Error())
				db.log.Printf("[ERROR] %s\n", msg)
				return errors.New(msg)
			}

		} else {
			defer func() {
				var err2 error
				if status {
					if err2 = tx.Commit(); err2 != nil {
						db.log.Printf("[ERROR] Failed to commit ad-hoc transaction: %s\n",
							err2.Error())
					}
				} else if err2 = tx.Rollback(); err2 != nil {
					db.log.Printf("[ERROR] Rollback of ad-hoc transaction failed: %s\n",
						err2.Error())
				}
			}()
		}
	}

	stmt = tx.Stmt(stmt)

EXEC_QUERY:
	if _, err = stmt.Exec(id); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("Cannot delete Search %d from database: %s",
				id,
				err.Error())
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	}

	status = true
	return nil
} // func (db *Database) SearchDelete(id int64) error

// SearchGetByID fetches a Search by its ID
func (db *Database) SearchGetByID(id int64) (*model.Search, error) {
	const qid query.ID = query.SearchGetByID
	var (
		err  error
		msg  string
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(id); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	if rows.Next() {
		var (
			timestamp  int64
			qstr, rstr string
			s          = &model.Search{ID: id}
		)

		if err = rows.Scan(&timestamp, &qstr, &rstr, &s.Count); err != nil {
			msg = fmt.Sprintf("Error scanning row for Host %d: %s",
				id,
				err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return nil, errors.New(msg)
		}

		if err = json.Unmarshal([]byte(qstr), &s.Query); err != nil {
			db.log.Printf("[ERROR] Cannot parse Query: %s\n\n%s\n\n",
				err.Error(),
				qstr)
			return nil, err
		} else if err = json.Unmarshal([]byte(rstr), &s.Results); err != nil {
			db.log.Printf("[ERROR] Cannot parse Results: %s\n\n%s\n\n",
				err.Error(),
				rstr)
			return nil, err
		}

		s.Timestamp = time.Unix(timestamp, 0)

		return s, nil
	}

	db.log.Printf("[INFO] Search #%d was not found in database\n", id)
	return nil, nil
} // func (db *Database) SearchGetByID(id int64) (*model.Search, error)

// SearchGetResults fetches the Records that were matched by a Search.
func (db *Database) SearchGetResults(id, offset, cnt int64) ([]model.Record, error) {
	const qid query.ID = query.SearchGetResults
	var (
		err  error
		msg  string
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(id, cnt, offset); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var records = make([]model.Record, 0)

	for rows.Next() {
		var (
			r         model.Record
			timestamp int64
		)

		if err = rows.Scan(&r.ID, &r.HostID, &timestamp, &r.Source, &r.Message); err != nil {
			msg = fmt.Sprintf("Failed to scan row: %s", err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return nil, errors.New(msg)
		}

		r.Time = time.Unix(timestamp, 0)
		records = append(records, r)
	}

	return records, nil
} // func (db *Database) SearchGetResults(id, offset, cnt int64) ([]model.Record, error)

// SearchGetAllID fetches the IDs and the number of results of all searches in the database.
func (db *Database) SearchGetAllID() ([][2]int64, error) {
	const qid query.ID = query.SearchGetAllID
	var (
		err  error
		msg  string
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return nil, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	var idlist = make([][2]int64, 0)

	for rows.Next() {
		var (
			id, cnt int64
		)

		if err = rows.Scan(&id, &cnt); err != nil {
			msg = fmt.Sprintf("Failed to scan row: %s", err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return nil, errors.New(msg)
		}

		idlist = append(idlist, [2]int64{id, cnt})
	}

	return idlist, nil
} // func (db *Database) SearchGetAllID() ([]int64, error)

// SearchGetResultCount returns the number of Records returned by the given Search.
func (db *Database) SearchGetResultCount(id int64) (int64, error) {
	const qid query.ID = query.SearchGetResultCount
	var (
		err  error
		msg  string
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return 0, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(id); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		}

		return 0, err
	}

	defer rows.Close() // nolint: errcheck,gosec

	if rows.Next() {
		var cnt int64

		if err = rows.Scan(&cnt); err != nil {
			msg = fmt.Sprintf("Error scanning row for Search %d: %s",
				id,
				err.Error())
			db.log.Printf("[ERROR] %s\n", msg)
			return 0, errors.New(msg)
		}

		return cnt, nil
	}

	return 0, fmt.Errorf("No Search with ID %d was found in the database", id)
} // func (db *Database) SearchGetResultCount(id int64) (int64, error)
