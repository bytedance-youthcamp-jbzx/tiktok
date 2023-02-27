// Package db /*
package db

import (
	"fmt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	"gorm.io/gorm/logger"
	"time"

	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var (
	_db       *gorm.DB
	config    = viper.Init("db")
	zapLogger = zap.InitLogger()
)

func getDsn(driverWithRole string) string {
	username := config.Viper.GetString(fmt.Sprintf("%s.username", driverWithRole))
	password := config.Viper.GetString(fmt.Sprintf("%s.password", driverWithRole))
	host := config.Viper.GetString(fmt.Sprintf("%s.host", driverWithRole))
	port := config.Viper.GetInt(fmt.Sprintf("%s.port", driverWithRole))
	Dbname := config.Viper.GetString(fmt.Sprintf("%s.database", driverWithRole))

	// data source name
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, Dbname)

	return dsn
}

func init() {
	zapLogger.Info("Redis server connection successful!")

	dsn1 := getDsn("mysql.source")

	var err error
	_db, err = gorm.Open(mysql.Open(dsn1), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Info),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(err.Error())
	}

	dsn2 := getDsn("mysql.replica1")
	dsn3 := getDsn("mysql.replica2")
	// 配置dbresolver
	_db.Use(dbresolver.Register(dbresolver.Config{
		// use `db1` as sources, `db2` as replicas
		Sources:  []gorm.Dialector{mysql.Open(dsn1)},
		Replicas: []gorm.Dialector{mysql.Open(dsn2), mysql.Open(dsn3)},
		// sources/replicas load balancing policy
		Policy: dbresolver.RandomPolicy{},
		// print sources/replicas mode in logger
		TraceResolverMode: false,
	}))
	// AutoMigrate会创建表，缺失的外键，约束，列和索引。如果大小，精度，是否为空，可以更改，则AutoMigrate会改变列的类型。出于保护您数据的目的，它不会删除未使用的列
	// 刷新数据库的表格，使其保持最新。即如果我在旧表的基础上增加一个字段age，那么调用autoMigrate后，旧表会自动多出一列age，值为空
	if err := _db.AutoMigrate(&User{}, &Video{}, &Comment{}, &FavoriteVideoRelation{}, &FollowRelation{}, &Message{}, &FavoriteCommentRelation{}); err != nil {
		zapLogger.Fatalln(err.Error())
	}

	db, err := _db.DB()
	if err != nil {
		zapLogger.Fatalln(err.Error())
	}
	db.SetMaxOpenConns(1000)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(60 * time.Minute)
}

func GetDB() *gorm.DB {
	return _db
}
