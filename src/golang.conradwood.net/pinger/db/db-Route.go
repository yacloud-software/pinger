package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBRoute
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence route_seq;

Main Table:

 CREATE TABLE route (id integer primary key default nextval('route_seq'),sourceip bigint not null  references ip (id) on delete cascade  ,destip bigint not null  references ip (id) on delete cascade  );

Alter statements:
ALTER TABLE route ADD COLUMN IF NOT EXISTS sourceip bigint not null references ip (id) on delete cascade  default 0;
ALTER TABLE route ADD COLUMN IF NOT EXISTS destip bigint not null references ip (id) on delete cascade  default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE route_archive (id integer unique not null,sourceip bigint not null,destip bigint not null);
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
	default_def_DBRoute *DBRoute
)

type DBRoute struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func DefaultDBRoute() *DBRoute {
	if default_def_DBRoute != nil {
		return default_def_DBRoute
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBRoute(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBRoute = res
	return res
}
func NewDBRoute(db *sql.DB) *DBRoute {
	foo := DBRoute{DB: db}
	foo.SQLTablename = "route"
	foo.SQLArchivetablename = "route_archive"
	return &foo
}

func (a *DBRoute) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBRoute) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBRoute) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBRoute) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBRoute", "insert into "+a.SQLArchivetablename+" (id,sourceip, destip) values ($1,$2, $3) ", p.ID, p.SourceIP.ID, p.DestIP.ID)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBRoute) buildSaveMap(ctx context.Context, p *savepb.Route) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["sourceip"] = a.get_col_from_proto(p, "sourceip")
	res["destip"] = a.get_col_from_proto(p, "destip")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBRoute) Save(ctx context.Context, p *savepb.Route) (uint64, error) {
	qn := "save_DBRoute"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBRoute) SaveWithID(ctx context.Context, p *savepb.Route) error {
	qn := "insert_DBRoute"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBRoute) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.Route) (uint64, error) {
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

func (a *DBRoute) Update(ctx context.Context, p *savepb.Route) error {
	qn := "DBRoute_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set sourceip=$1, destip=$2 where id = $3", a.get_SourceIP_ID(p), a.get_DestIP_ID(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBRoute) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBRoute_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBRoute) ByID(ctx context.Context, p uint64) (*savepb.Route, error) {
	qn := "DBRoute_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No Route with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) Route with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBRoute) TryByID(ctx context.Context, p uint64) (*savepb.Route, error) {
	qn := "DBRoute_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) Route with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBRoute) ByIDs(ctx context.Context, p []uint64) ([]*savepb.Route, error) {
	qn := "DBRoute_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBRoute) All(ctx context.Context) ([]*savepb.Route, error) {
	qn := "DBRoute_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBRoute" rows with matching SourceIP
func (a *DBRoute) BySourceIP(ctx context.Context, p uint64) ([]*savepb.Route, error) {
	qn := "DBRoute_BySourceIP"
	l, e := a.fromQuery(ctx, qn, "sourceip = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySourceIP: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRoute" rows with multiple matching SourceIP
func (a *DBRoute) ByMultiSourceIP(ctx context.Context, p []uint64) ([]*savepb.Route, error) {
	qn := "DBRoute_BySourceIP"
	l, e := a.fromQuery(ctx, qn, "sourceip in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySourceIP: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRoute) ByLikeSourceIP(ctx context.Context, p uint64) ([]*savepb.Route, error) {
	qn := "DBRoute_ByLikeSourceIP"
	l, e := a.fromQuery(ctx, qn, "sourceip ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("BySourceIP: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRoute" rows with matching DestIP
func (a *DBRoute) ByDestIP(ctx context.Context, p uint64) ([]*savepb.Route, error) {
	qn := "DBRoute_ByDestIP"
	l, e := a.fromQuery(ctx, qn, "destip = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDestIP: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRoute" rows with multiple matching DestIP
func (a *DBRoute) ByMultiDestIP(ctx context.Context, p []uint64) ([]*savepb.Route, error) {
	qn := "DBRoute_ByDestIP"
	l, e := a.fromQuery(ctx, qn, "destip in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDestIP: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRoute) ByLikeDestIP(ctx context.Context, p uint64) ([]*savepb.Route, error) {
	qn := "DBRoute_ByLikeDestIP"
	l, e := a.fromQuery(ctx, qn, "destip ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDestIP: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBRoute) get_ID(p *savepb.Route) uint64 {
	return uint64(p.ID)
}

// getter for reference "SourceIP"
func (a *DBRoute) get_SourceIP_ID(p *savepb.Route) uint64 {
	if p.SourceIP == nil {
		panic("field SourceIP must not be nil")
	}
	return p.SourceIP.ID
}

// getter for reference "DestIP"
func (a *DBRoute) get_DestIP_ID(p *savepb.Route) uint64 {
	if p.DestIP == nil {
		panic("field DestIP must not be nil")
	}
	return p.DestIP.ID
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBRoute) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.Route, error) {
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

func (a *DBRoute) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Route, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBRoute) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.Route, error) {
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
func (a *DBRoute) get_col_from_proto(p *savepb.Route, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "sourceip" {
		return a.get_SourceIP_ID(p)
	} else if colname == "destip" {
		return a.get_DestIP_ID(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBRoute) Tablename() string {
	return a.SQLTablename
}

func (a *DBRoute) SelectCols() string {
	return "id,sourceip, destip"
}
func (a *DBRoute) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".sourceip, " + a.SQLTablename + ".destip"
}

func (a *DBRoute) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.Route, error) {
	var res []*savepb.Route
	for rows.Next() {
		// SCANNER:
		foo := &savepb.Route{}
		// create the non-nullable pointers
		foo.SourceIP = &savepb.IP{} // non-nullable
		foo.DestIP = &savepb.IP{}   // non-nullable
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.SourceIP.ID
		scanTarget_2 := &foo.DestIP.ID
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
func (a *DBRoute) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),sourceip bigint not null ,destip bigint not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),sourceip bigint not null ,destip bigint not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS sourceip bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS destip bigint not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS sourceip bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS destip bigint not null  default 0;`,
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
		`ALTER TABLE ` + a.SQLTablename + ` add constraint mkdb_fk_route_sourceip_ipid FOREIGN KEY (sourceip) references ip (id) on delete cascade ;`,
		`ALTER TABLE ` + a.SQLTablename + ` add constraint mkdb_fk_route_destip_ipid FOREIGN KEY (destip) references ip (id) on delete cascade ;`,
	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBRoute) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

