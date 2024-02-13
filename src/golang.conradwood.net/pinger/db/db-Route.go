package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBRoute
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
	"golang.conradwood.net/go-easyops/sql"
	"os"
)

var (
	default_def_DBRoute *DBRoute
)

type DBRoute struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
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

// Save (and use database default ID generation)
func (a *DBRoute) Save(ctx context.Context, p *savepb.Route) (uint64, error) {
	qn := "DBRoute_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (sourceip, destip) values ($1, $2) returning id", p.SourceIP.ID, p.DestIP.ID)
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
func (a *DBRoute) SaveWithID(ctx context.Context, p *savepb.Route) error {
	qn := "insert_DBRoute"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,sourceip, destip) values ($1,$2, $3) ", p.ID, p.SourceIP.ID, p.DestIP.ID)
	return a.Error(ctx, qn, e)
}

func (a *DBRoute) Update(ctx context.Context, p *savepb.Route) error {
	qn := "DBRoute_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set sourceip=$1, destip=$2 where id = $3", p.SourceIP.ID, p.DestIP.ID, p.ID)

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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,sourceip, destip from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No Route with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Route with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBRoute) TryByID(ctx context.Context, p uint64) (*savepb.Route, error) {
	qn := "DBRoute_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,sourceip, destip from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Route with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBRoute) All(ctx context.Context) ([]*savepb.Route, error) {
	qn := "DBRoute_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,sourceip, destip from "+a.SQLTablename+" order by id")
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

// get all "DBRoute" rows with matching SourceIP
func (a *DBRoute) BySourceIP(ctx context.Context, p uint64) ([]*savepb.Route, error) {
	qn := "DBRoute_BySourceIP"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,sourceip, destip from "+a.SQLTablename+" where sourceip = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceIP: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceIP: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRoute) ByLikeSourceIP(ctx context.Context, p uint64) ([]*savepb.Route, error) {
	qn := "DBRoute_ByLikeSourceIP"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,sourceip, destip from "+a.SQLTablename+" where sourceip ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceIP: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySourceIP: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRoute" rows with matching DestIP
func (a *DBRoute) ByDestIP(ctx context.Context, p uint64) ([]*savepb.Route, error) {
	qn := "DBRoute_ByDestIP"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,sourceip, destip from "+a.SQLTablename+" where destip = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDestIP: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDestIP: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRoute) ByLikeDestIP(ctx context.Context, p uint64) ([]*savepb.Route, error) {
	qn := "DBRoute_ByLikeDestIP"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,sourceip, destip from "+a.SQLTablename+" where destip ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDestIP: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDestIP: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBRoute) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Route, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBRoute) Tablename() string {
	return a.SQLTablename
}

func (a *DBRoute) SelectCols() string {
	return "id,sourceip, destip"
}
func (a *DBRoute) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".sourceip, " + a.SQLTablename + ".destip"
}

func (a *DBRoute) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.Route, error) {
	var res []*savepb.Route
	for rows.Next() {
		foo := savepb.Route{SourceIP: &savepb.IP{}, DestIP: &savepb.IP{}}
		err := rows.Scan(&foo.ID, &foo.SourceIP.ID, &foo.DestIP.ID)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
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
		`ALTER TABLE route ADD COLUMN IF NOT EXISTS sourceip bigint not null default 0;`,
		`ALTER TABLE route ADD COLUMN IF NOT EXISTS destip bigint not null default 0;`,

		`ALTER TABLE route_archive ADD COLUMN IF NOT EXISTS sourceip bigint not null  default 0;`,
		`ALTER TABLE route_archive ADD COLUMN IF NOT EXISTS destip bigint not null  default 0;`,
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
		`ALTER TABLE route add constraint mkdb_fk_route_sourceip_ipid FOREIGN KEY (sourceip) references ip (id) on delete cascade ;`,
		`ALTER TABLE route add constraint mkdb_fk_route_destip_ipid FOREIGN KEY (destip) references ip (id) on delete cascade ;`,
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
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

