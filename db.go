package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

var db *sql.DB

var AddConfig = make(chan *Config)
var RemoveConfig = make(chan *Config)
var UpdateConfig = make(chan *Config)
var DbDone = make(chan struct{})

func DbOperaitons() {
	for {
		select {
		case config := <-AddConfig:
			err := config.Save()
			log.Println(err)
		case config := <-RemoveConfig:
			err := config.Remove()
			log.Println(err)
		case config := <-UpdateConfig:
			err := config.Update()
			log.Println(err)
		case <-DbDone:
			return
		}
	}
}

func LoadDataFromDb() {

}

func ResetDatabase() {
	os.Remove("./bouncer.db")
	var err error
	db, err = sql.Open("sqlite3", "./bouncer.db")
	if err != nil {
		log.Println(err)
	}
	defer db.Close()
	sqlStmt := `
    create table BackendServer (id integer not null primary key autoincrement, host text, config_id integer, foreign key(config_id) references Config);
    create table Config (id integer not null primary key autoincrement, host text, path text, targetPath text, reqPerSecond integer, maxConcurrentPerBackendServer integer);
    delete from Config;
    delete from BackendServer;
	`
	if _, err = db.Exec(sqlStmt); err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func (config *Config) Save() error {
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	stmt, err := tx.Prepare("insert into Config(host, path, targetPath, reqPerSecond, maxConcurrentPerBackendServer) values(?, ?, ?, ?, ?)")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(config.Host, config.Path, config.TargetPath, config.ReqPerSecond, config.MaxConcurrentPerBackendServer); err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}
	tx.Commit()

	tmpconfig, err := GetConfigDetails(config.Host, config.Path)
	if err != nil {
		log.Println(err)
		return err
	}

	tx, err = db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	stmt, err = tx.Prepare("delete from BackendServer where config_id = ?")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(tmpconfig.Id); err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}
	tx.Commit()

	for _, backendServer := range config.BackendServers {
		backendServer.ConfigId = tmpconfig.Id
		if err = backendServer.Save(); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil

}

func (config *Config) Update() error {
	err := config.Remove()
	if err != nil {
		return err
	}
	err = config.Save()
	if err != nil {
		return err
	}
	return nil
}

func (config *Config) Remove() error {
	configRemove, err := GetConfig(config.Host, config.Path)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	stmt, err := tx.Prepare("delete from BackendServer where config_id = ?")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(configRemove.Id); err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	stmt, err = tx.Prepare("delete from Config where config_id = ?")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(configRemove.Id); err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil

}

func (backendServer *BackendServer) Save() error {
	if backendServer.Id == 0 {

		tx, err := db.Begin()
		if err != nil {
			log.Println(err)
			return err
		}
		stmt, err := tx.Prepare("insert into BackendServer(host, config_id) values(?,?)")
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(backendServer.Host, backendServer.ConfigId)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil

	} else {

		tx, err := db.Begin()
		if err != nil {
			log.Println(err)
			return err
		}
		stmt, err := tx.Prepare("insert into BackendServer(id, host, config_id) values(?,?,?)")
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return err
		}
		defer stmt.Close()
		if _, err = stmt.Exec(backendServer.Id, backendServer.Host, backendServer.ConfigId); err != nil {
			log.Println(err)
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	}
}

func GetConfigDetails(host string, path string) (Config, error) {
	var config Config
	stmt, err := db.Prepare("select id, host, path, targetPath, reqPerSecond, maxConcurrentPerBackendServer from Config where host = ? and path = ?")
	if err != nil {
		log.Println(err)
		return config, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(host, path)
	var targetPath string
	var id int
	var reqPerSecond int
	var maxConcurrentPerBackendServer int
	err = row.Scan(&id, &host, &path, &targetPath, &reqPerSecond, &maxConcurrentPerBackendServer)
	config.Id = id
	config.Host = host
	config.Path = path
	config.TargetPath = targetPath
	config.ReqPerSecond = reqPerSecond
	config.MaxConcurrentPerBackendServer = maxConcurrentPerBackendServer
	return config, err
}

func GetConfig(host string, path string) (Config, error) {
	var config Config
	var err error
	config, err = GetConfigDetails(host, path)
	if err != nil {
		log.Println(err)
		return config, err
	}

	stmt, err := db.Prepare("select * from BackendServer where config_id = ?")
	if err != nil {
		log.Println(err)
		return config, err
	}
	defer stmt.Close()
	log.Println(config.Id)

	rows, err := stmt.Query(config.Id)
	if err != nil {
		log.Println(err)
		return config, err
	}
	defer rows.Close()
	for rows.Next() {
		var backendServer BackendServer
		if err = rows.Scan(&backendServer.Id, &backendServer.Host, &backendServer.ConfigId); err != nil {
			return config, err
		}
		config.BackendServers = append(config.BackendServers, backendServer)
	}
	return config, err
}
