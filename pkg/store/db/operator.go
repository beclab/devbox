package db

import (
	"fmt"
	"os"

	"github.com/beclab/devbox/pkg/store/db/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//
//const (
//	DefaultDBFile = "./data/message.db"
//)
//
//var (
//	dbFile = ""
//)
//
//func init() {
//	flag.StringVar(&dbFile, "db", DefaultDBFile, "default message db file")
//}

type DbOperator struct {
	DB *gorm.DB
}

var (
	db *gorm.DB
)

func init() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=allow",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	var err error
	//source := fmt.Sprintf("file:%s?cache=shared", dbFile)
	db, err = gorm.Open(postgres.New(
		postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: true,
		}),
		&gorm.Config{})
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
	} else {
		if !db.Migrator().HasColumn(&model.DevAppContainers{}, "Image") {
			err = db.Migrator().AddColumn(&model.DevAppContainers{}, "Image")
			if err != nil {
				return err
			}
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
