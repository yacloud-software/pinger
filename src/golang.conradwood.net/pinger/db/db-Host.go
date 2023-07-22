package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBHost
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
	"golang.conradwood.net/go-easyops/sql"
	"os"
)

var (
	default_def_DBHost *DBHost
)

type DBHost struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
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

// Save (and use database default ID generation)
func (a *DBHost) Save(ctx context.Context, p *savepb.Host) (uint64, error) {
	qn := "DBHost_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (name, pingerid) values ($1, $2) returning id", p.Name, p.PingerID)
	if e != nil {
		return 0, a.Error(ctx, qn, e)
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, a.Error(ctx, qn, fmt.Errorf("No rows after insert"))
	}
	var id uint64
	e = rows.Scan(&id)
	if e != nil {
		return 0, a.Error(ctx, qn, fmt.Errorf("failed to scan id after insert: %s", e))
	}
	p.ID = id
	return id, nil
}

// Save using the ID specified
func (a *DBHost) SaveWithID(ctx context.Context, p *savepb.Host) error {
	qn := "insert_DBHost"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,name, pingerid) values ($1,$2, $3) ", p.ID, p.Name, p.PingerID)
	return a.Error(ctx, qn, e)
}

func (a *DBHost) Update(ctx context.Context, p *savepb.Host) error {
	qn := "DBHost_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set name=$1, pingerid=$2 where id = $3", p.Name, p.PingerID, p.ID)

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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, pingerid from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No Host with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Host with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBHost) TryByID(ctx context.Context, p uint64) (*savepb.Host, error) {
	qn := "DBHost_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, pingerid from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("TryByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Host with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBHost) All(ctx context.Context) ([]*savepb.Host, error) {
	qn := "DBHost_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, pingerid from "+a.SQLTablename+" order by id")
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("All: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, fmt.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBHost" rows with matching Name
func (a *DBHost) ByName(ctx context.Context, p string) ([]*savepb.Host, error) {
	qn := "DBHost_ByName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, pingerid from "+a.SQLTablename+" where name = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBHost) ByLikeName(ctx context.Context, p string) ([]*savepb.Host, error) {
	qn := "DBHost_ByLikeName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, pingerid from "+a.SQLTablename+" where name ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBHost" rows with matching PingerID
func (a *DBHost) ByPingerID(ctx context.Context, p string) ([]*savepb.Host, error) {
	qn := "DBHost_ByPingerID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, pingerid from "+a.SQLTablename+" where pingerid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPingerID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPingerID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBHost) ByLikePingerID(ctx context.Context, p string) ([]*savepb.Host, error) {
	qn := "DBHost_ByLikePingerID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, pingerid from "+a.SQLTablename+" where pingerid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPingerID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPingerID: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBHost) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Host, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
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
		foo := savepb.Host{}
		err := rows.Scan(&foo.ID, &foo.Name, &foo.PingerID)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}

/**********************************************************************
* Helper to create table and columns
**********************************************************************/
func (a *DBHost) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),name text not null  ,pingerid text not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),name text not null  ,pingerid text not null  );`,
		`ALTER TABLE host ADD COLUMN IF NOT EXISTS name text not null default '';`,
		`ALTER TABLE host ADD COLUMN IF NOT EXISTS pingerid text not null default '';`,
	}
	for i, c := range csql {
		_, e := a.DB.ExecContext(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
		if e != nil {
			return e
		}
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
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
