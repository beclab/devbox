package db

import (
	"flag"
	"fmt"

	"github.com/beclab/devbox/pkg/store/db/model"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	DefaultDBFile = "/data/message.db"
)

var (
	dbFile = ""
)

func init() {
	flag.StringVar(&dbFile, "db", DefaultDBFile, "default message db file")
}

type DbOperator struct {
	DB *gorm.DB
}

var (
	db *gorm.DB
)

func init() {
	var err error
	source := fmt.Sprintf("file:%s?cache=shared", dbFile)
	db, err = gorm.Open(sqlite.Open(source), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = createTableIfNotExists()
	if err != nil {
		panic(err)
	}
}

func createTableIfNotExists() (err error) {
	if !db.Migrator().HasTable(model.DevApp{}) {
		err = db.Migrator().CreateTable(model.DevApp{})
		if err != nil {
			return err
		}
	}
	if !db.Migrator().HasTable(model.DevContainers{}) {
		err = db.Migrator().CreateTable(model.DevContainers{})
		if err != nil {
			return err
		}
	}
	if !db.Migrator().HasTable(model.DevAppContainers{}) {
		err = db.Migrator().CreateTable(model.DevAppContainers{})
		if err != nil {
			return err
		}
	}
	return nil
}

func NewDbOperator() *DbOperator {
	return &DbOperator{DB: db}
}

func (db *DbOperator) Close() error {
	return db.Close()
}
