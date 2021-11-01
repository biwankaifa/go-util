package db

import (
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"net/url"
	"time"
)

type MysqlConnectPool struct{}

type MysqlConfig struct {
	User            string // 账号
	Pass            string // 密码
	Addr            string // 服务器地址及端口
	Name            string // 表名
	MaxOpenConn     int    // 同时打开的连接数
	MaxIdleConn     int    // 链接池中最多保留的空闲链接
	ConnMaxLifeTime int    // 可重用链接得最大时间长度
	RunMode         string // 允许模式
}

var db *gorm.DB

var instance *MysqlConnectPool

// Get 获取对象
// Deprecated: 已废弃请使用 Db() 方法
func Get() *MysqlConnectPool {
	return instance
}

// Db 获取对象
func Db() *gorm.DB {
	return db
}

// InitMysql 初始化数据库连接
func (c *MysqlConfig) InitMysql() error {
	var err error
	db, err = c.dbConnect()
	if err != nil {
		return err
	}
	return nil
}

// dbConnect
func (c *MysqlConfig) dbConnect() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=%t&loc=%s&time_zone=%s",
		c.User,
		c.Pass,
		c.Addr,
		c.Name,
		true,
		"Local",
		url.QueryEscape("'Asia/Shanghai'"),
	)

	var loggerConfig logger.Interface
	if c.RunMode == "debug" {
		loggerConfig = logger.Default.LogMode(logger.Info)
	} else {
		loggerConfig = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: loggerConfig, // 日志配置
	})

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("[db connection failed] Database name: %s", c.Name))
	}

	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。
	sqlDB.SetMaxOpenConns(c.MaxOpenConn)

	// 设置最大连接数 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(c.MaxIdleConn)

	// 设置最大连接超时
	sqlDB.SetConnMaxLifetime(time.Minute * time.Duration(c.ConnMaxLifeTime))

	// 使用插件
	//db.Use(&TracePlugin{})

	return db, nil
}
