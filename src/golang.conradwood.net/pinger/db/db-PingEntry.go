package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBPingEntry
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence pingentry_seq;

Main Table:

 CREATE TABLE pingentry (id integer primary key default nextval('pingentry_seq'),ip text not null  ,interval integer not null  ,metrichostname text not null  ,pingerid text not null  ,label text not null  ,ipversion integer not null  ,isactive boolean not null  );

Alter statements:
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS ip text not null default '';
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS interval integer not null default 0;
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS metrichostname text not null default '';
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS pingerid text not null default '';
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS label text not null default '';
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS ipversion integer not null default 0;
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS isactive boolean not null default false;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE pingentry_archive (id integer unique not null,ip text not null,interval integer not null,metrichostname text not null,pingerid text not null,label text not null,ipversion integer not null,isactive boolean not null);
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
	default_def_DBPingEntry *DBPingEntry
)

type DBPingEntry struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBPingEntry() *DBPingEntry {
	if default_def_DBPingEntry != nil {
		return default_def_DBPingEntry
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBPingEntry(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBPingEntry = res
	return res
}
func NewDBPingEntry(db *sql.DB) *DBPingEntry {
	foo := DBPingEntry{DB: db}
	foo.SQLTablename = "pingentry"
	foo.SQLArchivetablename = "pingentry_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBPingEntry) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBPingEntry", "insert into "+a.SQLArchivetablename+" (id,ip, interval, metrichostname, pingerid, label, ipversion, isactive) values ($1,$2, $3, $4, $5, $6, $7, $8) ", p.ID, p.IP, p.Interval, p.MetricHostName, p.PingerID, p.Label, p.IPVersion, p.IsActive)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBPingEntry) Save(ctx context.Context, p *savepb.PingEntry) (uint64, error) {
	qn := "DBPingEntry_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (ip, interval, metrichostname, pingerid, label, ipversion, isactive) values ($1, $2, $3, $4, $5, $6, $7) returning id", p.IP, p.Interval, p.MetricHostName, p.PingerID, p.Label, p.IPVersion, p.IsActive)
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
func (a *DBPingEntry) SaveWithID(ctx context.Context, p *savepb.PingEntry) error {
	qn := "insert_DBPingEntry"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,ip, interval, metrichostname, pingerid, label, ipversion, isactive) values ($1,$2, $3, $4, $5, $6, $7, $8) ", p.ID, p.IP, p.Interval, p.MetricHostName, p.PingerID, p.Label, p.IPVersion, p.IsActive)
	return a.Error(ctx, qn, e)
}

func (a *DBPingEntry) Update(ctx context.Context, p *savepb.PingEntry) error {
	qn := "DBPingEntry_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set ip=$1, interval=$2, metrichostname=$3, pingerid=$4, label=$5, ipversion=$6, isactive=$7 where id = $8", p.IP, p.Interval, p.MetricHostName, p.PingerID, p.Label, p.IPVersion, p.IsActive, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBPingEntry) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBPingEntry_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBPingEntry) ByID(ctx context.Context, p uint64) (*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No PingEntry with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) PingEntry with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBPingEntry) TryByID(ctx context.Context, p uint64) (*savepb.PingEntry, error) {
	qn := "DBPingEntry_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) PingEntry with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBPingEntry) All(ctx context.Context) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" order by id")
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

// get all "DBPingEntry" rows with matching IP
func (a *DBPingEntry) ByIP(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByIP"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where ip = $1", p)
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
func (a *DBPingEntry) ByLikeIP(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeIP"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where ip ilike $1", p)
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

// get all "DBPingEntry" rows with matching Interval
func (a *DBPingEntry) ByInterval(ctx context.Context, p uint32) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByInterval"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where interval = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInterval: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInterval: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeInterval(ctx context.Context, p uint32) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeInterval"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where interval ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInterval: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByInterval: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with matching MetricHostName
func (a *DBPingEntry) ByMetricHostName(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByMetricHostName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where metrichostname = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMetricHostName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMetricHostName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeMetricHostName(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeMetricHostName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where metrichostname ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMetricHostName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByMetricHostName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with matching PingerID
func (a *DBPingEntry) ByPingerID(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByPingerID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where pingerid = $1", p)
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
func (a *DBPingEntry) ByLikePingerID(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikePingerID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where pingerid ilike $1", p)
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

// get all "DBPingEntry" rows with matching Label
func (a *DBPingEntry) ByLabel(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLabel"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where label = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLabel: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLabel: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeLabel(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeLabel"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where label ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLabel: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLabel: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with matching IPVersion
func (a *DBPingEntry) ByIPVersion(ctx context.Context, p uint32) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByIPVersion"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where ipversion = $1", p)
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
func (a *DBPingEntry) ByLikeIPVersion(ctx context.Context, p uint32) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeIPVersion"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where ipversion ilike $1", p)
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

// get all "DBPingEntry" rows with matching IsActive
func (a *DBPingEntry) ByIsActive(ctx context.Context, p bool) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByIsActive"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where isactive = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByIsActive: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByIsActive: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeIsActive(ctx context.Context, p bool) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeIsActive"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,ip, interval, metrichostname, pingerid, label, ipversion, isactive from "+a.SQLTablename+" where isactive ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByIsActive: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByIsActive: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBPingEntry) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.PingEntry, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBPingEntry) Tablename() string {
	return a.SQLTablename
}

func (a *DBPingEntry) SelectCols() string {
	return "id,ip, interval, metrichostname, pingerid, label, ipversion, isactive"
}
func (a *DBPingEntry) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".ip, " + a.SQLTablename + ".interval, " + a.SQLTablename + ".metrichostname, " + a.SQLTablename + ".pingerid, " + a.SQLTablename + ".label, " + a.SQLTablename + ".ipversion, " + a.SQLTablename + ".isactive"
}

func (a *DBPingEntry) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.PingEntry, error) {
	var res []*savepb.PingEntry
	for rows.Next() {
		foo := savepb.PingEntry{}
		err := rows.Scan(&foo.ID, &foo.IP, &foo.Interval, &foo.MetricHostName, &foo.PingerID, &foo.Label, &foo.IPVersion, &foo.IsActive)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}
func (a *DBPingEntry) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.PingEntry, error) {
	var res []*savepb.PingEntry
	for rows.Next() {
		// SCANNER:
		foo := &savepb.PingEntry{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.IP
		scanTarget_2 := &foo.Interval
		scanTarget_3 := &foo.MetricHostName
		scanTarget_4 := &foo.PingerID
		scanTarget_5 := &foo.Label
		scanTarget_6 := &foo.IPVersion
		scanTarget_7 := &foo.IsActive
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5, scanTarget_6, scanTarget_7)
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
func (a *DBPingEntry) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),ip text not null ,interval integer not null ,metrichostname text not null ,pingerid text not null ,label text not null ,ipversion integer not null ,isactive boolean not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),ip text not null ,interval integer not null ,metrichostname text not null ,pingerid text not null ,label text not null ,ipversion integer not null ,isactive boolean not null );`,
		`ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS ip text not null default '';`,
		`ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS interval integer not null default 0;`,
		`ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS metrichostname text not null default '';`,
		`ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS pingerid text not null default '';`,
		`ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS label text not null default '';`,
		`ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS ipversion integer not null default 0;`,
		`ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS isactive boolean not null default false;`,

		`ALTER TABLE pingentry_archive ADD COLUMN IF NOT EXISTS ip text not null  default '';`,
		`ALTER TABLE pingentry_archive ADD COLUMN IF NOT EXISTS interval integer not null  default 0;`,
		`ALTER TABLE pingentry_archive ADD COLUMN IF NOT EXISTS metrichostname text not null  default '';`,
		`ALTER TABLE pingentry_archive ADD COLUMN IF NOT EXISTS pingerid text not null  default '';`,
		`ALTER TABLE pingentry_archive ADD COLUMN IF NOT EXISTS label text not null  default '';`,
		`ALTER TABLE pingentry_archive ADD COLUMN IF NOT EXISTS ipversion integer not null  default 0;`,
		`ALTER TABLE pingentry_archive ADD COLUMN IF NOT EXISTS isactive boolean not null  default false;`,
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
func (a *DBPingEntry) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

