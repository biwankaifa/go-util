package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database

type ConfigOfMongo struct {
	Host     string
	Port     int
	Name     string
	Pass     string
	DataBase string
	Str      string
}

// Db 获取对象
func Db() *mongo.Database {
	return db
}

// InitMongo 初始化数据库连接
func (cfg *ConfigOfMongo) InitMongo() error {
	if db != nil {
		return nil
	}
	var param string
	if cfg.Str == "" {
		param = fmt.Sprintf("mongodb://%s:%s@%s:%d", cfg.Name, cfg.Pass, cfg.Host, cfg.Port)
	} else {
		param = cfg.Str
	}
	clientOptions := options.Client().ApplyURI(param)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}

	db = client.Database(cfg.DataBase)

	return nil
}
