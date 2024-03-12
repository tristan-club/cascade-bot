package dbutil

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/tristan-club/kit/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	rawLog "log"
	"os"
	"strings"
	"time"
)

const (
	baseLink             = "{db_user}:{db_password}@tcp({db_host}:{db_port})/{db_name}?charset=utf8mb4&parseTime=true"
	defaultMigrationPath = "./db/migrations"
	defaultMigrationEnv  = "DB_MIGRATION_PATH"
)

const (
	DB_HOST     = "DB_HOST"
	DB_PORT     = "DB_PORT"
	DB_NAME     = "DB_NAME"
	DB_USER     = "DB_USER"
	DB_PASSWORD = "DB_PASSWORD"
)

var db *gorm.DB = nil

type DaoMysql struct {
}

func NewDaoMysql() *DaoMysql {
	return &DaoMysql{}
}

type MysqlConnection struct {
	*gorm.DB
	IsRead bool
}

func InitDB(migrateSignal migrate.MigrationDirection) error {
	// connect to chain_info database
	dbHost := os.Getenv(DB_HOST)
	dbPort := os.Getenv(DB_PORT)
	dbName := os.Getenv(DB_NAME)
	dbUser := os.Getenv(DB_USER)
	dbPassword := os.Getenv(DB_PASSWORD)
	err := ConnectDb(dbHost, dbPort, dbName, dbUser, dbPassword)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action": "connect to db",
			"error":  err.Error(),
		})
		return err
	}

	if migrateSignal == migrate.Up {
		if err := doMigration(); err != nil {
			return err
		}
	}

	return nil
}
func doMigration() error {

	//if config.EnvIsDev() {
	//	return nil
	//}

	migrationPath := os.Getenv(defaultMigrationEnv)
	if migrationPath == "" {
		migrationPath = defaultMigrationPath
	}
	migrations := &migrate.FileMigrationSource{
		Dir: migrationPath,
	}
	Orm := GetDb(nil)
	sqlDb, err := Orm.DB()
	if err != nil {
		log.Error().Msgf("sqlMigrate Orm.DB.DB() err ", err)
		return err
	}
	code, err := migrate.Exec(sqlDb, "mysql", migrations, migrate.Up)
	if err != nil {
		log.Error().Msgf("sqlMigrate err %s", err)
		return err
	}
	log.Info().Msgf("migrate tabel num: %d", code)

	return nil

}

func ConnectDb(dbHost string, dbPort string, dbName string, dbUser string, dbPassword string) error {
	currentLink := buildLinkUrl(dbHost, dbPort, dbName, dbUser, dbPassword)
	var err error
	mysqlConfig := mysql.New(mysql.Config{
		DSN:                       currentLink,
		SkipInitializeWithVersion: false,
	})

	conf := &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}

	if os.Getenv("DB_LOG") == "1" {
		conf.Logger = glogger.New(
			rawLog.New(os.Stdout, "\r\n", rawLog.LstdFlags),
			glogger.Config{
				SlowThreshold: time.Second,
				LogLevel:      glogger.Info,
				Colorful:      true,
			})
	}

	db, err = gorm.Open(mysqlConfig, conf)
	if err != nil {
		log.Error().Msgf("failed to connect mysql cause:%s", err.Error())
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(16)
	sqlDB.SetMaxOpenConns(160)
	sqlDB.SetConnMaxLifetime(time.Hour)
	log.Info().Fields(map[string]interface{}{
		"action": "connect to db",
		"host":   dbHost,
		"port":   dbPort,
		"dbName": dbName,
	}).Send()
	return nil
}

func buildLinkUrl(dbHost string, dbPort string, dbName string, dbUser string, dbPassword string) string {
	currentLink := baseLink
	currentLink = strings.Replace(currentLink, "{db_user}", dbUser, -1)
	currentLink = strings.Replace(currentLink, "{db_password}", dbPassword, -1)
	currentLink = strings.Replace(currentLink, "{db_host}", dbHost, -1)
	currentLink = strings.Replace(currentLink, "{db_port}", dbPort, -1)
	currentLink = strings.Replace(currentLink, "{db_name}", dbName, -1)
	return currentLink
}

func Default() *gorm.DB {
	return db
}
func GetDb(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return db
}

type JSON json.RawMessage

func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}
