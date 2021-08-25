package db

import (
	"fmt"
	"go-util/config"
	"gorm.io/gorm/logger"
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type MysqlConnectPool struct {
}

var instance *MysqlConnectPool
var once sync.Once

var db *gorm.DB
var errDb error

func Get() *MysqlConnectPool {
	once.Do(func() {
		instance = &MysqlConnectPool{}
	})
	return instance
}

//InitDataPool 初始化数据库连接(可在mail()适当位置调用)
func (m *MysqlConnectPool) InitDataPool() bool {

	cfg := config.Get().MySQL

	db, errDb = dbConnect(cfg.Write.User, cfg.Write.Pass, cfg.Write.Addr, cfg.Write.Name)
	if errDb != nil {
		log.Fatal(errDb)
		return false
	}
	//关闭数据库，db会被多个goroutine共享，可以不调用
	// defer db.Close()
	return true
}

// Db 对外获取数据库连接对象db
func (m *MysqlConnectPool) Db() (db_con *gorm.DB) {
	return db
}

// Error 对外获取数据库连接对象db
func (m *MysqlConnectPool) Error() (error error) {
	return errDb
}

// dbConnect
func dbConnect(user, pass, addr, dbName string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=%t&loc=%s",
		user,
		pass,
		addr,
		dbName,
		true,
		"Local")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logger.Info), // 日志配置
	})

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("[db connection failed] Database name: %s", dbName))
	}

	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	cfg := config.Get().MySQL.Base

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)

	// 设置最大连接数 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)

	// 设置最大连接超时
	sqlDB.SetConnMaxLifetime(time.Minute * cfg.ConnMaxLifeTime)

	// 使用插件
	//db.Use(&TracePlugin{})

	return db, nil
}
