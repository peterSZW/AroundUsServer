package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/inconshreveable/log15"
	_ "github.com/mattn/go-sqlite3"
)

//savecode,email=logic_readmailcode(req.Name,req.Email)
// if savecode==req.Code {
// 	login_link(req.Name,req.email)
// }
//savecode,email=logic_readmailcode(req.Name,req.Email)
// if savecode==req.Code {
// 	login_link(req.Name,req.email)
// }

func runSQL(db *sql.DB, sqlcmd string) error {

	stmt, err := db.Prepare(sqlcmd)
	if !ErrIsNil_LOG(err, sqlcmd) {
		return err
	}
	_, err = stmt.Exec()
	if !ErrIsNil_LOG(err, sqlcmd) {
		return err
	}
	return nil
}

const DBFILE = "./server.db"

func init() {
	var err error
	db, err := sql.Open("sqlite3", DBFILE)
	checkErr(err)
	defer db.Close()

	// stmt, err = db.Prepare("delete from userinfo where uid=?")
	// checkErr(err)
	// res, err = stmt.Exec(id)
	// checkErr(err)
	runSQL(db, `CREATE TABLE users(
		uuid     CHAR(50),
		create_time TIMESTAMP NOT NULL DEFAULT current_timestamp,
		UNIQUE ( uuid ) 
	)`)

	runSQL(db, `CREATE TABLE email_code(
			email     CHAR(50),
			code      CHAR(50),
			name      char(50),
			create_time TIMESTAMP NOT NULL DEFAULT current_timestamp,
			UNIQUE ( email,name ) 
		)`)
	runSQL(db, `CREATE TABLE user_email(
			name  char(50),
			email CHAR(50),
			UNIQUE ( name )
		)`)

	// stmt, err := db.Prepare(`
	// CREATE TABLE email_code(
	// 	email     CHAR(50),
	// 	code        CHAR(50),
	// 	name char(50),
	// 	create_time TIMESTAMP NOT NULL DEFAULT current_timestamp,
	// 	UNIQUE ( email,name )
	//  )`)
	// if ErrIsNil_LOG(err, "Create email_code table") {
	// 	_, err = stmt.Exec()
	// 	ErrIsNil_LOG(err, "Create email_code table exec")
	// }

	// fmt.Println("================================")

	// stmt, err = db.Prepare(`
	// CREATE TABLE user_email(
	// 	name char(50),
	// 	email CHAR(50),
	// 	UNIQUE ( name )
	//  )`)
	// if ErrIsNil_LOG(err, "Create user_email table") {
	// 	_, err = stmt.Exec()
	// 	ErrIsNil_LOG(err, "Create user_email table exec")
	// }

	// fmt.Println("================================")
}

/*
1. gen code...
2. save code
3. check code,and  link name
4. expire code
info
  logic_mail_getemail(name)
*/
func logic_mail_expirecode() error {
	db, err := sql.Open("sqlite3", DBFILE)
	if !ErrIsNil_LOG(err, "open db") {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare("delete from email_code where create_time<datetime('now','-24 hour')")
	if !ErrIsNil_LOG(err, "delete email_code Prepare") {
		return err
	}
	_, err = stmt.Exec()
	if !ErrIsNil_LOG(err, "delete email_code exec") {
		return err
	}
	return nil

}
func logic_mail_getemail(name string) string {
	db, err := sql.Open("sqlite3", DBFILE)
	if !ErrIsNil_LOG(err, "open db") {
		return ""
	}
	defer db.Close()

	rows, err := db.Query("SELECT email  FROM user_email  where name=?", name)
	if !ErrIsNil_LOG(err, "select rows") {
		return ""
	}

	fmt.Println(rows)
	var email string

	for rows.Next() {
		err = rows.Scan(&email)
		if !ErrIsNil_LOG(err, "rows.scan") {
			return ""
		}

	}
	rows.Close()
	return email
}

func logic_mail_savecode(name, email, code string) error {

	// try update, if not success,try insert
	db, err := sql.Open("sqlite3", DBFILE)
	if !ErrIsNil_LOG(err, "open db") {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare("update email_code set code=?,create_time= current_timestamp,email=? where name=?")
	if !ErrIsNil_LOG(err, "update email_code Prepare") {
		return err
	}
	res, err := stmt.Exec(code, email, name)
	if !ErrIsNil_LOG(err, "update email_code exec") {
		return err
	}
	affect, err := res.RowsAffected()
	if !ErrIsNil_LOG(err, "rows") {
		return err
	}
	fmt.Println(affect)
	if affect == 0 {
		stmt, err := db.Prepare("insert into  email_code(name,email,code) values(?,?,?)")
		if !ErrIsNil_LOG(err, "Insert email_code Prepare") {
			return err
		}
		res, err := stmt.Exec(name, email, code)
		if !ErrIsNil_LOG(err, "insert email_code exec") {
			return err
		}
		affect2, err := res.RowsAffected()
		if !ErrIsNil_LOG(err, "rows") {
			return err
		}
		fmt.Println(affect2)

		if affect2 != 1 {
			log15.Error("insert email_code not ok")
			return errors.New("insert email_code affect!=1")
		}
	}

	return nil
}

func logic_mail_checkcode(code string) error {

	db, err := sql.Open("sqlite3", DBFILE)
	if !ErrIsNil_LOG(err, "open db") {
		return err
	}
	defer db.Close()
	// try update, if not success,try insert

	rows, err := db.Query("SELECT email,name FROM email_code where code=?", code)
	if !ErrIsNil_LOG(err, "select rows") {
		return err
	}

	fmt.Println(rows)
	var email string
	var name string

	for rows.Next() {
		err = rows.Scan(&email, &name)
		if !ErrIsNil_LOG(err, "rows.scan") {
			return err
		}

	}

	rows.Close()
	if email != "" {
		fmt.Println(email, name)

		err = logic_mail_link(db, name, email)
		return err
	}

	return nil
}

func logic_mail_link(db *sql.DB, name, email string) error {
	// db, err := sql.Open("sqlite3", DBFILE)
	// if !ErrIsNil_LOG(err, "open db") {
	// 	return err
	// }
	// defer db.Close()

	stmt, err := db.Prepare("update user_email set email=?  where name=?")
	if !ErrIsNil_LOG(err, "update user_email Prepare") {
		return err
	}
	res, err := stmt.Exec(email, name)
	if !ErrIsNil_LOG(err, "update user_email exec") {
		return err
	}
	affect, err := res.RowsAffected()
	if !ErrIsNil_LOG(err, "rows") {
		return err
	}
	if affect == 0 {
		stmt, err := db.Prepare("insert into user_email(name,email) values(?,?)")
		if !ErrIsNil_LOG(err, "Insert user_email Prepare") {
			return err
		}
		res, err := stmt.Exec(name, email)
		if !ErrIsNil_LOG(err, "insert user_email exec") {
			return err
		}
		affect2, err := res.RowsAffected()
		if !ErrIsNil_LOG(err, "insert user_email rows") {
			return err
		}
		if affect2 != 1 {
			log15.Error("insert user_email affect!=1")
			return errors.New("insert user_email affect!=1")
		}
	}

	return nil

}

func logic_mail_linkx() {

	//删除数据
	// stmt, err = db.Prepare("delete from userinfo where uid=?")
	// checkErr(err)

	// res, err = stmt.Exec(id)
	// checkErr(err)

	// affect, err = res.RowsAffected()
	// checkErr(err)

	// fmt.Println(affect)

	//db.Close()
}

func ErrIsNil_LOG(err error, logmsg string) bool {
	//fmt.Println(logmsg, "======", err)
	if err != nil {
		log15.Error(logmsg, "err", err)
	}
	return (err == nil)
}

func checkErr(err error) {

	if err != nil {
		log15.Error("checkErr", "err", err)

		panic(err)
	}
}

func logic_get_uuids() []string {
	var uuids []string
	db, err := sql.Open("sqlite3", DBFILE)
	if !ErrIsNil_LOG(err, "open db") {
		return uuids
	}
	defer db.Close()

	rows, err := db.Query("SELECT uuid  FROM users ")
	if !ErrIsNil_LOG(err, "select rows") {
		return uuids
	}

	fmt.Println(rows)

	for rows.Next() {
		var auuid string
		err = rows.Scan(&auuid)
		if !ErrIsNil_LOG(err, "rows.scan") {
			return uuids
		}
		uuids = append(uuids, auuid)

	}
	rows.Close()
	return uuids
}

func logic_get_uuid(uuid string) string {
	db, err := sql.Open("sqlite3", DBFILE)
	if !ErrIsNil_LOG(err, "open db") {
		return ""
	}
	defer db.Close()

	rows, err := db.Query("SELECT uuid  FROM users  where uuid=?", uuid)
	if !ErrIsNil_LOG(err, "select rows") {
		return ""
	}

	fmt.Println(rows)
	var uuids string

	for rows.Next() {
		err = rows.Scan(&uuids)
		if !ErrIsNil_LOG(err, "rows.scan") {
			return ""
		}

	}
	rows.Close()
	return uuids
}

func logic_save_uuid(uuid string) error {

	// try update, if not success,try insert
	db, err := sql.Open("sqlite3", DBFILE)
	if !ErrIsNil_LOG(err, "open db") {
		return err
	}
	defer db.Close()

	sql := "insert into  users(uuid) values(?)"
	stmt, err := db.Prepare(sql)
	if !ErrIsNil_LOG(err, sql) {
		return err
	}
	res, err := stmt.Exec(uuid)
	if !ErrIsNil_LOG(err, sql+"(Exec)") {
		return err
	}
	affect2, err := res.RowsAffected()
	if !ErrIsNil_LOG(err, "rows") {
		return err
	}
	fmt.Println(affect2)
	return nil

}

func logic_delete_uuid(uuid string) error {

	// try update, if not success,try insert
	db, err := sql.Open("sqlite3", DBFILE)
	if !ErrIsNil_LOG(err, "open db") {
		return err
	}
	defer db.Close()

	sql := "delete from   users where uuid=?"
	stmt, err := db.Prepare(sql)
	if !ErrIsNil_LOG(err, sql) {
		return err
	}
	res, err := stmt.Exec(uuid)
	if !ErrIsNil_LOG(err, sql+"(Exec)") {
		return err
	}
	affect2, err := res.RowsAffected()
	if !ErrIsNil_LOG(err, "rows") {
		return err
	}
	fmt.Println(affect2)
	return nil

}
