package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBPingEntry
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence pingentry_seq;

Main Table:

 CREATE TABLE pingentry (id integer primary key default nextval('pingentry_seq'),ip text not null  ,interval integer not null  ,metrichostname text not null  ,pingerid text not null  ,label text not null  ,ipversion integer not null  ,isactive boolean not null  ,label2 text not null  ,label3 text not null  ,label4 text not null  );

Alter statements:
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS ip text not null default '';
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS interval integer not null default 0;
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS metrichostname text not null default '';
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS pingerid text not null default '';
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS label text not null default '';
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS ipversion integer not null default 0;
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS isactive boolean not null default false;
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS label2 text not null default '';
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS label3 text not null default '';
ALTER TABLE pingentry ADD COLUMN IF NOT EXISTS label4 text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE pingentry_archive (id integer unique not null,ip text not null,interval integer not null,metrichostname text not null,pingerid text not null,label text not null,ipversion integer not null,isactive boolean not null,label2 text not null,label3 text not null,label4 text not null);
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
	default_def_DBPingEntry *DBPingEntry
)

type DBPingEntry struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
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

func (a *DBPingEntry) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBPingEntry) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBPingEntry) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBPingEntry) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBPingEntry", "insert into "+a.SQLArchivetablename+" (id,ip, interval, metrichostname, pingerid, label, ipversion, isactive, label2, label3, label4) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ", p.ID, p.IP, p.Interval, p.MetricHostName, p.PingerID, p.Label, p.IPVersion, p.IsActive, p.Label2, p.Label3, p.Label4)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBPingEntry) buildSaveMap(ctx context.Context, p *savepb.PingEntry) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["ip"] = a.get_col_from_proto(p, "ip")
	res["interval"] = a.get_col_from_proto(p, "interval")
	res["metrichostname"] = a.get_col_from_proto(p, "metrichostname")
	res["pingerid"] = a.get_col_from_proto(p, "pingerid")
	res["label"] = a.get_col_from_proto(p, "label")
	res["ipversion"] = a.get_col_from_proto(p, "ipversion")
	res["isactive"] = a.get_col_from_proto(p, "isactive")
	res["label2"] = a.get_col_from_proto(p, "label2")
	res["label3"] = a.get_col_from_proto(p, "label3")
	res["label4"] = a.get_col_from_proto(p, "label4")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBPingEntry) Save(ctx context.Context, p *savepb.PingEntry) (uint64, error) {
	qn := "save_DBPingEntry"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBPingEntry) SaveWithID(ctx context.Context, p *savepb.PingEntry) error {
	qn := "insert_DBPingEntry"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBPingEntry) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.PingEntry) (uint64, error) {
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

func (a *DBPingEntry) Update(ctx context.Context, p *savepb.PingEntry) error {
	qn := "DBPingEntry_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set ip=$1, interval=$2, metrichostname=$3, pingerid=$4, label=$5, ipversion=$6, isactive=$7, label2=$8, label3=$9, label4=$10 where id = $11", a.get_IP(p), a.get_Interval(p), a.get_MetricHostName(p), a.get_PingerID(p), a.get_Label(p), a.get_IPVersion(p), a.get_IsActive(p), a.get_Label2(p), a.get_Label3(p), a.get_Label4(p), p.ID)

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
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No PingEntry with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) PingEntry with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBPingEntry) TryByID(ctx context.Context, p uint64) (*savepb.PingEntry, error) {
	qn := "DBPingEntry_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) PingEntry with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBPingEntry) ByIDs(ctx context.Context, p []uint64) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBPingEntry) All(ctx context.Context) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBPingEntry" rows with matching IP
func (a *DBPingEntry) ByIP(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByIP"
	l, e := a.fromQuery(ctx, qn, "ip = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIP: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with multiple matching IP
func (a *DBPingEntry) ByMultiIP(ctx context.Context, p []string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByIP"
	l, e := a.fromQuery(ctx, qn, "ip in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIP: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeIP(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeIP"
	l, e := a.fromQuery(ctx, qn, "ip ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIP: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with matching Interval
func (a *DBPingEntry) ByInterval(ctx context.Context, p uint32) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByInterval"
	l, e := a.fromQuery(ctx, qn, "interval = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByInterval: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with multiple matching Interval
func (a *DBPingEntry) ByMultiInterval(ctx context.Context, p []uint32) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByInterval"
	l, e := a.fromQuery(ctx, qn, "interval in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByInterval: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeInterval(ctx context.Context, p uint32) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeInterval"
	l, e := a.fromQuery(ctx, qn, "interval ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByInterval: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with matching MetricHostName
func (a *DBPingEntry) ByMetricHostName(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByMetricHostName"
	l, e := a.fromQuery(ctx, qn, "metrichostname = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMetricHostName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with multiple matching MetricHostName
func (a *DBPingEntry) ByMultiMetricHostName(ctx context.Context, p []string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByMetricHostName"
	l, e := a.fromQuery(ctx, qn, "metrichostname in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMetricHostName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeMetricHostName(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeMetricHostName"
	l, e := a.fromQuery(ctx, qn, "metrichostname ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByMetricHostName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with matching PingerID
func (a *DBPingEntry) ByPingerID(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByPingerID"
	l, e := a.fromQuery(ctx, qn, "pingerid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPingerID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with multiple matching PingerID
func (a *DBPingEntry) ByMultiPingerID(ctx context.Context, p []string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByPingerID"
	l, e := a.fromQuery(ctx, qn, "pingerid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPingerID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikePingerID(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikePingerID"
	l, e := a.fromQuery(ctx, qn, "pingerid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByPingerID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with matching Label
func (a *DBPingEntry) ByLabel(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLabel"
	l, e := a.fromQuery(ctx, qn, "label = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLabel: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with multiple matching Label
func (a *DBPingEntry) ByMultiLabel(ctx context.Context, p []string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLabel"
	l, e := a.fromQuery(ctx, qn, "label in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLabel: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeLabel(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeLabel"
	l, e := a.fromQuery(ctx, qn, "label ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLabel: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with matching IPVersion
func (a *DBPingEntry) ByIPVersion(ctx context.Context, p uint32) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByIPVersion"
	l, e := a.fromQuery(ctx, qn, "ipversion = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIPVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with multiple matching IPVersion
func (a *DBPingEntry) ByMultiIPVersion(ctx context.Context, p []uint32) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByIPVersion"
	l, e := a.fromQuery(ctx, qn, "ipversion in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIPVersion: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeIPVersion(ctx context.Context, p uint32) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeIPVersion"
	l, e := a.fromQuery(ctx, qn, "ipversion ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIPVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with matching IsActive
func (a *DBPingEntry) ByIsActive(ctx context.Context, p bool) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByIsActive"
	l, e := a.fromQuery(ctx, qn, "isactive = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIsActive: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with multiple matching IsActive
func (a *DBPingEntry) ByMultiIsActive(ctx context.Context, p []bool) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByIsActive"
	l, e := a.fromQuery(ctx, qn, "isactive in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIsActive: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeIsActive(ctx context.Context, p bool) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeIsActive"
	l, e := a.fromQuery(ctx, qn, "isactive ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIsActive: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with matching Label2
func (a *DBPingEntry) ByLabel2(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLabel2"
	l, e := a.fromQuery(ctx, qn, "label2 = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLabel2: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with multiple matching Label2
func (a *DBPingEntry) ByMultiLabel2(ctx context.Context, p []string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLabel2"
	l, e := a.fromQuery(ctx, qn, "label2 in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLabel2: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeLabel2(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeLabel2"
	l, e := a.fromQuery(ctx, qn, "label2 ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLabel2: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with matching Label3
func (a *DBPingEntry) ByLabel3(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLabel3"
	l, e := a.fromQuery(ctx, qn, "label3 = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLabel3: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with multiple matching Label3
func (a *DBPingEntry) ByMultiLabel3(ctx context.Context, p []string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLabel3"
	l, e := a.fromQuery(ctx, qn, "label3 in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLabel3: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeLabel3(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeLabel3"
	l, e := a.fromQuery(ctx, qn, "label3 ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLabel3: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with matching Label4
func (a *DBPingEntry) ByLabel4(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLabel4"
	l, e := a.fromQuery(ctx, qn, "label4 = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLabel4: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingEntry" rows with multiple matching Label4
func (a *DBPingEntry) ByMultiLabel4(ctx context.Context, p []string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLabel4"
	l, e := a.fromQuery(ctx, qn, "label4 in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLabel4: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingEntry) ByLikeLabel4(ctx context.Context, p string) ([]*savepb.PingEntry, error) {
	qn := "DBPingEntry_ByLikeLabel4"
	l, e := a.fromQuery(ctx, qn, "label4 ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLabel4: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBPingEntry) get_ID(p *savepb.PingEntry) uint64 {
	return uint64(p.ID)
}

// getter for field "IP" (IP) [string]
func (a *DBPingEntry) get_IP(p *savepb.PingEntry) string {
	return string(p.IP)
}

// getter for field "Interval" (Interval) [uint32]
func (a *DBPingEntry) get_Interval(p *savepb.PingEntry) uint32 {
	return uint32(p.Interval)
}

// getter for field "MetricHostName" (MetricHostName) [string]
func (a *DBPingEntry) get_MetricHostName(p *savepb.PingEntry) string {
	return string(p.MetricHostName)
}

// getter for field "PingerID" (PingerID) [string]
func (a *DBPingEntry) get_PingerID(p *savepb.PingEntry) string {
	return string(p.PingerID)
}

// getter for field "Label" (Label) [string]
func (a *DBPingEntry) get_Label(p *savepb.PingEntry) string {
	return string(p.Label)
}

// getter for field "IPVersion" (IPVersion) [uint32]
func (a *DBPingEntry) get_IPVersion(p *savepb.PingEntry) uint32 {
	return uint32(p.IPVersion)
}

// getter for field "IsActive" (IsActive) [bool]
func (a *DBPingEntry) get_IsActive(p *savepb.PingEntry) bool {
	return bool(p.IsActive)
}

// getter for field "Label2" (Label2) [string]
func (a *DBPingEntry) get_Label2(p *savepb.PingEntry) string {
	return string(p.Label2)
}

// getter for field "Label3" (Label3) [string]
func (a *DBPingEntry) get_Label3(p *savepb.PingEntry) string {
	return string(p.Label3)
}

// getter for field "Label4" (Label4) [string]
func (a *DBPingEntry) get_Label4(p *savepb.PingEntry) string {
	return string(p.Label4)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBPingEntry) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.PingEntry, error) {
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

func (a *DBPingEntry) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.PingEntry, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBPingEntry) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.PingEntry, error) {
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
func (a *DBPingEntry) get_col_from_proto(p *savepb.PingEntry, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "ip" {
		return a.get_IP(p)
	} else if colname == "interval" {
		return a.get_Interval(p)
	} else if colname == "metrichostname" {
		return a.get_MetricHostName(p)
	} else if colname == "pingerid" {
		return a.get_PingerID(p)
	} else if colname == "label" {
		return a.get_Label(p)
	} else if colname == "ipversion" {
		return a.get_IPVersion(p)
	} else if colname == "isactive" {
		return a.get_IsActive(p)
	} else if colname == "label2" {
		return a.get_Label2(p)
	} else if colname == "label3" {
		return a.get_Label3(p)
	} else if colname == "label4" {
		return a.get_Label4(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBPingEntry) Tablename() string {
	return a.SQLTablename
}

func (a *DBPingEntry) SelectCols() string {
	return "id,ip, interval, metrichostname, pingerid, label, ipversion, isactive, label2, label3, label4"
}
func (a *DBPingEntry) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".ip, " + a.SQLTablename + ".interval, " + a.SQLTablename + ".metrichostname, " + a.SQLTablename + ".pingerid, " + a.SQLTablename + ".label, " + a.SQLTablename + ".ipversion, " + a.SQLTablename + ".isactive, " + a.SQLTablename + ".label2, " + a.SQLTablename + ".label3, " + a.SQLTablename + ".label4"
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
		scanTarget_8 := &foo.Label2
		scanTarget_9 := &foo.Label3
		scanTarget_10 := &foo.Label4
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5, scanTarget_6, scanTarget_7, scanTarget_8, scanTarget_9, scanTarget_10)
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
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),ip text not null ,interval integer not null ,metrichostname text not null ,pingerid text not null ,label text not null ,ipversion integer not null ,isactive boolean not null ,label2 text not null ,label3 text not null ,label4 text not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),ip text not null ,interval integer not null ,metrichostname text not null ,pingerid text not null ,label text not null ,ipversion integer not null ,isactive boolean not null ,label2 text not null ,label3 text not null ,label4 text not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS ip text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS interval integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS metrichostname text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS pingerid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS label text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS ipversion integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS isactive boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS label2 text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS label3 text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS label4 text not null default '';`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS ip text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS interval integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS metrichostname text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS pingerid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS label text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS ipversion integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS isactive boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS label2 text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS label3 text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS label4 text not null  default '';`,
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
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

