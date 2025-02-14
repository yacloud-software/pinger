package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBHost
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence host_seq;

Main Table:

 CREATE TABLE host (id integer primary key default nextval('host_seq'),name text not null  ,pingerid text not null  );

Alter statements:
ALTER TABLE host ADD COLUMN IF NOT EXISTS name text not null default '';
ALTER TABLE host ADD COLUMN IF NOT EXISTS pingerid text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE host_archive (id integer unique not null,name text not null,pingerid text not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/sql"
	"os"
	"sync"
)

var (
	default_def_DBHost *DBHost
)

type DBHost struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func DefaultDBHost() *DBHost {
	if default_def_DBHost != nil {
		return default_def_DBHost
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBHost(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBHost = res
	return res
}
func NewDBHost(db *sql.DB) *DBHost {
	foo := DBHost{DB: db}
	foo.SQLTablename = "host"
	foo.SQLArchivetablename = "host_archive"
	return &foo
}

func (a *DBHost) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBHost) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBHost) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBHost) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBHost", "insert into "+a.SQLArchivetablename+" (id,name, pingerid) values ($1,$2, $3) ", p.ID, p.Name, p.PingerID)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBHost) buildSaveMap(ctx context.Context, p *savepb.Host) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["name"] = a.get_col_from_proto(p, "name")
	res["pingerid"] = a.get_col_from_proto(p, "pingerid")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBHost) Save(ctx context.Context, p *savepb.Host) (uint64, error) {
	qn := "save_DBHost"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBHost) SaveWithID(ctx context.Context, p *savepb.Host) error {
	qn := "insert_DBHost"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBHost) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.Host) (uint64, error) {
	// Save (and use database default ID generation)

	var rows *gosql.Rows
	var e error

	q_cols := ""
	q_valnames := ""
	q_vals := make([]interface{}, 0)
	deli := ""
	i := 0
	// build the 2 parts of the query (column names and value names) as well as the values themselves
	for colname, val := range smap {
		q_cols = q_cols + deli + colname
		i++
		q_valnames = q_valnames + deli + fmt.Sprintf("$%d", i)
		q_vals = append(q_vals, val)
		deli = ","
	}
	rows, e = a.DB.QueryContext(ctx, queryname, "insert into "+a.SQLTablename+" ("+q_cols+") values ("+q_valnames+") returning id", q_vals...)
	if e != nil {
		return 0, a.Error(ctx, queryname, e)
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, a.Error(ctx, queryname, errors.Errorf("No rows after insert"))
	}
	var id uint64
	e = rows.Scan(&id)
	if e != nil {
		return 0, a.Error(ctx, queryname, errors.Errorf("failed to scan id after insert: %s", e))
	}
	p.ID = id
	return id, nil
}

func (a *DBHost) Update(ctx context.Context, p *savepb.Host) error {
	qn := "DBHost_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set name=$1, pingerid=$2 where id = $3", a.get_Name(p), a.get_PingerID(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBHost) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBHost_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBHost) ByID(ctx context.Context, p uint64) (*savepb.Host, error) {
	qn := "DBHost_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No Host with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) Host with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBHost) TryByID(ctx context.Context, p uint64) (*savepb.Host, error) {
	qn := "DBHost_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) Host with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBHost) ByIDs(ctx context.Context, p []uint64) ([]*savepb.Host, error) {
	qn := "DBHost_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBHost) All(ctx context.Context) ([]*savepb.Host, error) {
	qn := "DBHost_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBHost" rows with matching Name
func (a *DBHost) ByName(ctx context.Context, p string) ([]*savepb.Host, error) {
	qn := "DBHost_ByName"
	l, e := a.fromQuery(ctx, qn, "name = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBHost" rows with multiple matching Name
func (a *DBHost) ByMultiName(ctx context.Context, p []string) ([]*savepb.Host, error) {
	qn := "DBHost_ByName"
	l, e := a.fromQuery(ctx, qn, "name in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBHost) ByLikeName(ctx context.Context, p string) ([]*savepb.Host, error) {
	qn := "DBHost_ByLikeName"
	l, e := a.fromQuery(ctx, qn, "name ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBHost" rows with matching PingerID
func (a *DBHost) ByPingerID(ctx context.Context, p string) ([]*savepb.Host, error) {
	qn := "DBHost_ByPingerID"
	l, e := a.fromQuery(ctx, qn, "pingerid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPingerID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBHost" rows with multiple matching PingerID
func (a *DBHost) ByMultiPingerID(ctx context.Context, p []string) ([]*savepb.Host, error) {
	qn := "DBHost_ByPingerID"
	l, e := a.fromQuery(ctx, qn, "pingerid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPingerID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBHost) ByLikePingerID(ctx context.Context, p string) ([]*savepb.Host, error) {
	qn := "DBHost_ByLikePingerID"
	l, e := a.fromQuery(ctx, qn, "pingerid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPingerID: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBHost) get_ID(p *savepb.Host) uint64 {
	return uint64(p.ID)
}

// getter for field "Name" (Name) [string]
func (a *DBHost) get_Name(p *savepb.Host) string {
	return string(p.Name)
}

// getter for field "PingerID" (PingerID) [string]
func (a *DBHost) get_PingerID(p *savepb.Host) string {
	return string(p.PingerID)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBHost) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.Host, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	i := 0
	for col_name, value := range extra_fields {
		i++
		efname := fmt.Sprintf("EXTRA_FIELD_%d", i)
		query.Add(col_name+" = "+efname, QP{efname: value})
	}

	gw, paras := query.ToPostgres()
	queryname := "custom_dbquery"
	rows, err := a.DB.QueryContext(ctx, queryname, "select "+a.SelectCols()+" from "+a.Tablename()+" where "+gw, paras...)
	if err != nil {
		return nil, err
	}
	res, err := a.FromRows(ctx, rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	return res, nil

}

func (a *DBHost) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Host, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBHost) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.Host, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	eq := ""
	if extra_fields != nil && len(extra_fields) > 0 {
		eq = " AND ("
		// build the extraquery "eq"
		i := len(args)
		deli := ""
		for col_name, value := range extra_fields {
			i++
			eq = eq + deli + col_name + fmt.Sprintf(" = $%d", i)
			deli = " AND "
			args = append(args, value)
		}
		eq = eq + ")"
	}
	rows, err := a.DB.QueryContext(ctx, queryname, "select "+a.SelectCols()+" from "+a.Tablename()+" where ( "+query_where+") "+eq, args...)
	if err != nil {
		return nil, err
	}
	res, err := a.FromRows(ctx, rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	return res, nil
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBHost) get_col_from_proto(p *savepb.Host, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "name" {
		return a.get_Name(p)
	} else if colname == "pingerid" {
		return a.get_PingerID(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBHost) Tablename() string {
	return a.SQLTablename
}

func (a *DBHost) SelectCols() string {
	return "id,name, pingerid"
}
func (a *DBHost) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".name, " + a.SQLTablename + ".pingerid"
}

func (a *DBHost) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.Host, error) {
	var res []*savepb.Host
	for rows.Next() {
		// SCANNER:
		foo := &savepb.Host{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.Name
		scanTarget_2 := &foo.PingerID
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2)
		// END SCANNER

		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, foo)
	}
	return res, nil
}

/**********************************************************************
* Helper to create table and columns
**********************************************************************/
func (a *DBHost) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),name text not null ,pingerid text not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),name text not null ,pingerid text not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS name text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS pingerid text not null default '';`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS name text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS pingerid text not null  default '';`,
	}

	for i, c := range csql {
		_, e := a.DB.ExecContext(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
		if e != nil {
			return e
		}
	}

	// these are optional, expected to fail
	csql = []string{
		// Indices:

		// Foreign keys:

	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBHost) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

