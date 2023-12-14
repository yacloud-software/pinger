package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBIP
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence ip_seq;

Main Table:

 CREATE TABLE ip (id integer primary key default nextval('ip_seq'),ipversion integer not null  ,name text not null  ,ip text not null  ,host bigint not null  references host (id) on delete cascade  );

Alter statements:
ALTER TABLE ip ADD COLUMN IF NOT EXISTS ipversion integer not null default 0;
ALTER TABLE ip ADD COLUMN IF NOT EXISTS name text not null default '';
ALTER TABLE ip ADD COLUMN IF NOT EXISTS ip text not null default '';
ALTER TABLE ip ADD COLUMN IF NOT EXISTS host bigint not null references host (id) on delete cascade  default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE ip_archive (id integer unique not null,ipversion integer not null,name text not null,ip text not null,host bigint not null);
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
	default_def_DBIP *DBIP
)

type DBIP struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBIP() *DBIP {
	if default_def_DBIP != nil {
		return default_def_DBIP
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBIP(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBIP = res
	return res
}
func NewDBIP(db *sql.DB) *DBIP {
	foo := DBIP{DB: db}
	foo.SQLTablename = "ip"
	foo.SQLArchivetablename = "ip_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBIP) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBIP", "insert into "+a.SQLArchivetablename+" (id,ipversion, name, ip, host) values ($1,$2, $3, $4, $5) ", p.ID, p.IPVersion, p.Name, p.IP, p.Host.ID)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBIP) Save(ctx context.Context, p *savepb.IP) (uint64, error) {
	qn := "DBIP_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (ipversion, name, ip, host) values ($1, $2, $3, $4) returning id", p.IPVersion, p.Name, p.IP, p.Host.ID)
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
func (a *DBIP) SaveWithID(ctx context.Context, p *savepb.IP) error {
	qn := "insert_DBIP"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,ipversion, name, ip, host) values ($1,$2, $3, $4, $5) ", p.ID, p.IPVersion, p.Name, p.IP, p.Host.ID)
	return a.Error(ctx, qn, e)
}

func (a *DBIP) Update(ctx context.Context, p *savepb.IP) error {
	qn := "DBIP_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set ipversion=$1, name=$2, ip=$3, host=$4 where id = $5", p.IPVersion, p.Name, p.IP, p.Host.ID, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBIP) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBIP_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBIP) ByID(ctx context.Context, p uint64) (*savepb.IP, error) {
	qn := "DBIP_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ipversion, name, ip, host from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No IP with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) IP with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBIP) TryByID(ctx context.Context, p uint64) (*savepb.IP, error) {
	qn := "DBIP_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ipversion, name, ip, host from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) IP with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBIP) All(ctx context.Context) ([]*savepb.IP, error) {
	qn := "DBIP_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ipversion, name, ip, host from "+a.SQLTablename+" order by id")
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

// get all "DBIP" rows with matching IPVersion
func (a *DBIP) ByIPVersion(ctx context.Context, p uint32) ([]*savepb.IP, error) {
	qn := "DBIP_ByIPVersion"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ipversion, name, ip, host from "+a.SQLTablename+" where ipversion = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByIPVersion: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByIPVersion: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBIP) ByLikeIPVersion(ctx context.Context, p uint32) ([]*savepb.IP, error) {
	qn := "DBIP_ByLikeIPVersion"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ipversion, name, ip, host from "+a.SQLTablename+" where ipversion ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByIPVersion: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByIPVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBIP" rows with matching Name
func (a *DBIP) ByName(ctx context.Context, p string) ([]*savepb.IP, error) {
	qn := "DBIP_ByName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ipversion, name, ip, host from "+a.SQLTablename+" where name = $1", p)
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
func (a *DBIP) ByLikeName(ctx context.Context, p string) ([]*savepb.IP, error) {
	qn := "DBIP_ByLikeName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ipversion, name, ip, host from "+a.SQLTablename+" where name ilike $1", p)
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

// get all "DBIP" rows with matching IP
func (a *DBIP) ByIP(ctx context.Context, p string) ([]*savepb.IP, error) {
	qn := "DBIP_ByIP"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ipversion, name, ip, host from "+a.SQLTablename+" where ip = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByIP: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByIP: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBIP) ByLikeIP(ctx context.Context, p string) ([]*savepb.IP, error) {
	qn := "DBIP_ByLikeIP"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ipversion, name, ip, host from "+a.SQLTablename+" where ip ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByIP: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByIP: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBIP" rows with matching Host
func (a *DBIP) ByHost(ctx context.Context, p uint64) ([]*savepb.IP, error) {
	qn := "DBIP_ByHost"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ipversion, name, ip, host from "+a.SQLTablename+" where host = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByHost: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBIP) ByLikeHost(ctx context.Context, p uint64) ([]*savepb.IP, error) {
	qn := "DBIP_ByLikeHost"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ipversion, name, ip, host from "+a.SQLTablename+" where host ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByHost: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBIP) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.IP, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBIP) Tablename() string {
	return a.SQLTablename
}

func (a *DBIP) SelectCols() string {
	return "id,ipversion, name, ip, host"
}
func (a *DBIP) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".ipversion, " + a.SQLTablename + ".name, " + a.SQLTablename + ".ip, " + a.SQLTablename + ".host"
}

func (a *DBIP) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.IP, error) {
	var res []*savepb.IP
	for rows.Next() {
		foo := savepb.IP{Host: &savepb.Host{}}
		err := rows.Scan(&foo.ID, &foo.IPVersion, &foo.Name, &foo.IP, &foo.Host.ID)
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
func (a *DBIP) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),ipversion integer not null  ,name text not null  ,ip text not null  ,host bigint not null  references host (id) on delete cascade  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),ipversion integer not null  ,name text not null  ,ip text not null  ,host bigint not null  references host (id) on delete cascade  );`,
		`ALTER TABLE ip ADD COLUMN IF NOT EXISTS ipversion integer not null default 0;`,
		`ALTER TABLE ip ADD COLUMN IF NOT EXISTS name text not null default '';`,
		`ALTER TABLE ip ADD COLUMN IF NOT EXISTS ip text not null default '';`,
		`ALTER TABLE ip ADD COLUMN IF NOT EXISTS host bigint not null references host (id) on delete cascade  default 0;`,
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
func (a *DBIP) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}






