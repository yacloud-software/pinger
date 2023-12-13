package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBTag
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence tag_seq;

Main Table:

 CREATE TABLE tag (id integer primary key default nextval('tag_seq'),tagname text not null  );

Alter statements:
ALTER TABLE tag ADD COLUMN IF NOT EXISTS tagname text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE tag_archive (id integer unique not null,tagname text not null);
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
	default_def_DBTag *DBTag
)

type DBTag struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBTag() *DBTag {
	if default_def_DBTag != nil {
		return default_def_DBTag
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBTag(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBTag = res
	return res
}
func NewDBTag(db *sql.DB) *DBTag {
	foo := DBTag{DB: db}
	foo.SQLTablename = "tag"
	foo.SQLArchivetablename = "tag_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBTag) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBTag", "insert into "+a.SQLArchivetablename+" (id,tagname) values ($1,$2) ", p.ID, p.TagName)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBTag) Save(ctx context.Context, p *savepb.Tag) (uint64, error) {
	qn := "DBTag_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (tagname) values ($1) returning id", p.TagName)
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
func (a *DBTag) SaveWithID(ctx context.Context, p *savepb.Tag) error {
	qn := "insert_DBTag"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,tagname) values ($1,$2) ", p.ID, p.TagName)
	return a.Error(ctx, qn, e)
}

func (a *DBTag) Update(ctx context.Context, p *savepb.Tag) error {
	qn := "DBTag_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set tagname=$1 where id = $2", p.TagName, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBTag) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBTag_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBTag) ByID(ctx context.Context, p uint64) (*savepb.Tag, error) {
	qn := "DBTag_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,tagname from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No Tag with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Tag with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBTag) TryByID(ctx context.Context, p uint64) (*savepb.Tag, error) {
	qn := "DBTag_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,tagname from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Tag with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBTag) All(ctx context.Context) ([]*savepb.Tag, error) {
	qn := "DBTag_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,tagname from "+a.SQLTablename+" order by id")
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

// get all "DBTag" rows with matching TagName
func (a *DBTag) ByTagName(ctx context.Context, p string) ([]*savepb.Tag, error) {
	qn := "DBTag_ByTagName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,tagname from "+a.SQLTablename+" where tagname = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTagName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTagName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBTag) ByLikeTagName(ctx context.Context, p string) ([]*savepb.Tag, error) {
	qn := "DBTag_ByLikeTagName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,tagname from "+a.SQLTablename+" where tagname ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTagName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTagName: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBTag) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Tag, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBTag) Tablename() string {
	return a.SQLTablename
}

func (a *DBTag) SelectCols() string {
	return "id,tagname"
}
func (a *DBTag) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".tagname"
}

func (a *DBTag) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.Tag, error) {
	var res []*savepb.Tag
	for rows.Next() {
		foo := savepb.Tag{}
		err := rows.Scan(&foo.ID, &foo.TagName)
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
func (a *DBTag) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),tagname text not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),tagname text not null  );`,
		`ALTER TABLE tag ADD COLUMN IF NOT EXISTS tagname text not null default '';`,
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
func (a *DBTag) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}




