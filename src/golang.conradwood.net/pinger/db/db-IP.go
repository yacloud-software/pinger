package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBIP
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
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/sql"
	"os"
	"sync"
)

var (
	default_def_DBIP *DBIP
)

type DBIP struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBIP()
	})
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

func (a *DBIP) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBIP) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBIP) NewQuery() *Query {
	return newQuery(a)
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

// return a map with columnname -> value_from_proto
func (a *DBIP) buildSaveMap(ctx context.Context, p *savepb.IP) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["ipversion"] = a.get_col_from_proto(p, "ipversion")
	res["name"] = a.get_col_from_proto(p, "name")
	res["ip"] = a.get_col_from_proto(p, "ip")
	res["host"] = a.get_col_from_proto(p, "host")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBIP) Save(ctx context.Context, p *savepb.IP) (uint64, error) {
	qn := "save_DBIP"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBIP) SaveWithID(ctx context.Context, p *savepb.IP) error {
	qn := "insert_DBIP"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBIP) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.IP) (uint64, error) {
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

// if ID==0 save, otherwise update
func (a *DBIP) SaveOrUpdate(ctx context.Context, p *savepb.IP) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBIP) Update(ctx context.Context, p *savepb.IP) error {
	qn := "DBIP_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set ipversion=$1, name=$2, ip=$3, host=$4 where id = $5", a.get_IPVersion(p), a.get_Name(p), a.get_IP(p), a.get_Host_ID(p), p.ID)

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
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No IP with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) IP with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBIP) TryByID(ctx context.Context, p uint64) (*savepb.IP, error) {
	qn := "DBIP_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) IP with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBIP) ByIDs(ctx context.Context, p []uint64) ([]*savepb.IP, error) {
	qn := "DBIP_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBIP) All(ctx context.Context) ([]*savepb.IP, error) {
	qn := "DBIP_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBIP" rows with matching IPVersion
func (a *DBIP) ByIPVersion(ctx context.Context, p uint32) ([]*savepb.IP, error) {
	qn := "DBIP_ByIPVersion"
	l, e := a.fromQuery(ctx, qn, "ipversion = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIPVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBIP" rows with multiple matching IPVersion
func (a *DBIP) ByMultiIPVersion(ctx context.Context, p []uint32) ([]*savepb.IP, error) {
	qn := "DBIP_ByIPVersion"
	l, e := a.fromQuery(ctx, qn, "ipversion in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIPVersion: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBIP) ByLikeIPVersion(ctx context.Context, p uint32) ([]*savepb.IP, error) {
	qn := "DBIP_ByLikeIPVersion"
	l, e := a.fromQuery(ctx, qn, "ipversion ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIPVersion: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBIP" rows with matching Name
func (a *DBIP) ByName(ctx context.Context, p string) ([]*savepb.IP, error) {
	qn := "DBIP_ByName"
	l, e := a.fromQuery(ctx, qn, "name = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBIP" rows with multiple matching Name
func (a *DBIP) ByMultiName(ctx context.Context, p []string) ([]*savepb.IP, error) {
	qn := "DBIP_ByName"
	l, e := a.fromQuery(ctx, qn, "name in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBIP) ByLikeName(ctx context.Context, p string) ([]*savepb.IP, error) {
	qn := "DBIP_ByLikeName"
	l, e := a.fromQuery(ctx, qn, "name ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBIP" rows with matching IP
func (a *DBIP) ByIP(ctx context.Context, p string) ([]*savepb.IP, error) {
	qn := "DBIP_ByIP"
	l, e := a.fromQuery(ctx, qn, "ip = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIP: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBIP" rows with multiple matching IP
func (a *DBIP) ByMultiIP(ctx context.Context, p []string) ([]*savepb.IP, error) {
	qn := "DBIP_ByIP"
	l, e := a.fromQuery(ctx, qn, "ip in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIP: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBIP) ByLikeIP(ctx context.Context, p string) ([]*savepb.IP, error) {
	qn := "DBIP_ByLikeIP"
	l, e := a.fromQuery(ctx, qn, "ip ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByIP: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBIP" rows with matching Host
func (a *DBIP) ByHost(ctx context.Context, p uint64) ([]*savepb.IP, error) {
	qn := "DBIP_ByHost"
	l, e := a.fromQuery(ctx, qn, "host = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBIP" rows with multiple matching Host
func (a *DBIP) ByMultiHost(ctx context.Context, p []uint64) ([]*savepb.IP, error) {
	qn := "DBIP_ByHost"
	l, e := a.fromQuery(ctx, qn, "host in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBIP) ByLikeHost(ctx context.Context, p uint64) ([]*savepb.IP, error) {
	qn := "DBIP_ByLikeHost"
	l, e := a.fromQuery(ctx, qn, "host ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBIP) get_ID(p *savepb.IP) uint64 {
	return uint64(p.ID)
}

// getter for field "IPVersion" (IPVersion) [uint32]
func (a *DBIP) get_IPVersion(p *savepb.IP) uint32 {
	return uint32(p.IPVersion)
}

// getter for field "Name" (Name) [string]
func (a *DBIP) get_Name(p *savepb.IP) string {
	return string(p.Name)
}

// getter for field "IP" (IP) [string]
func (a *DBIP) get_IP(p *savepb.IP) string {
	return string(p.IP)
}

// getter for reference "Host"
func (a *DBIP) get_Host_ID(p *savepb.IP) uint64 {
	if p.Host == nil {
		panic("field Host must not be nil")
	}
	return p.Host.ID
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBIP) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.IP, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	i := 0
	for col_name, value := range extra_fields {
		i++
		/*
		   efname:=fmt.Sprintf("EXTRA_FIELD_%d",i)
		   query.Add(col_name+" = "+efname,QP{efname:value})
		*/
		query.AddEqual(col_name, value)
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

func (a *DBIP) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.IP, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBIP) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.IP, error) {
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
func (a *DBIP) get_col_from_proto(p *savepb.IP, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "ipversion" {
		return a.get_IPVersion(p)
	} else if colname == "name" {
		return a.get_Name(p)
	} else if colname == "ip" {
		return a.get_IP(p)
	} else if colname == "host" {
		return a.get_Host_ID(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

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
		// SCANNER:
		foo := &savepb.IP{}
		// create the non-nullable pointers
		foo.Host = &savepb.Host{} // non-nullable
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.IPVersion
		scanTarget_2 := &foo.Name
		scanTarget_3 := &foo.IP
		scanTarget_4 := &foo.Host.ID
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4)
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
func (a *DBIP) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),ipversion integer not null ,name text not null ,ip text not null ,host bigint not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),ipversion integer not null ,name text not null ,ip text not null ,host bigint not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS ipversion integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS name text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS ip text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS host bigint not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS ipversion integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS name text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS ip text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS host bigint not null  default 0;`,
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
		`ALTER TABLE ` + a.SQLTablename + ` add constraint mkdb_fk_ip_host_hostid FOREIGN KEY (host) references host (id) on delete cascade ;`,
	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
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
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

