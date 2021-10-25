package config

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"strings"
	"time"

	consulApi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

type Config struct {
	// source config获取来源方式, file 本地文件方式获取, consul 从consul的kv内获取
	source string
	// address consul下地址获取
	address string
	// path kv路径信息
	path string
	// configType kv数据格式
	configType string
}

func (c *Config) SetSource(value string) *Config {
	c.source = value
	return c
}

func (c *Config) SetAddress(value string) *Config {
	c.address = value
	return c
}

func (c *Config) SetPath(value string) *Config {
	c.path = value
	return c
}

func (c *Config) SetType(value string) *Config {
	c.configType = value
	return c
}

var DefaultConfig *viper.Viper

func initConsulConfig(address, path, configType string) *viper.Viper {
	//consulAddress = "http://127.0.0.1:8500"
	//consulPath = "config"

	DefaultConfig = viper.New()
	DefaultConfig.SetConfigType(configType)

	consulClient, err := consulApi.NewClient(&consulApi.Config{Address: address})
	if err != nil {
		log.Fatalln("consul连接失败:", err)
	}

	kv, _, err := consulClient.KV().Get(path, nil)
	if err != nil {
		log.Fatalln("consul获取配置失败:", err)
	}

	err = DefaultConfig.ReadConfig(bytes.NewBuffer(kv.Value))
	if err != nil {
		log.Fatalln("Viper解析配置失败:", err)
	}

	go func() {
		time.Sleep(time.Second * 10)
		params := make(map[string]interface{})
		params["type"] = "key"
		params["key"] = path

		w, err := watch.Parse(params)
		if err != nil {
			log.Fatalln(err)
		}
		w.Handler = func(u uint64, i interface{}) {
			kv := i.(*consulApi.KVPair)
			hotconfig := viper.New()
			hotconfig.SetConfigType(configType)
			err = hotconfig.ReadConfig(bytes.NewBuffer(kv.Value))
			if err != nil {
				log.Fatalln("Viper解析配置失败:", err)
			}
			DefaultConfig = hotconfig
		}
		err = w.Run(address)
		if err != nil {
			log.Fatalln("监听consul错误:", err)
		}
	}()

	return DefaultConfig
}

func initFileConfig(path, configType string) *viper.Viper {
	DefaultConfig = viper.New()
	DefaultConfig.SetConfigName("configs")
	DefaultConfig.SetConfigType(configType)
	DefaultConfig.AddConfigPath("./" + path)

	if err := DefaultConfig.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("config.toml not found")
		} else {
			panic(err)
		}
	}

	//创建一个信道等待关闭
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		//设置监听回调函数
		DefaultConfig.OnConfigChange(func(e fsnotify.Event) {
			fmt.Printf("config is change :%s \n", e.String())
			//if err := viper.Unmarshal(config); err != nil {
			//	panic(err)
			//}
			cancel()
		})
		//开始监听
		DefaultConfig.WatchConfig()
		//信道不会主动关闭，可以主动调用cancel关闭
		<-ctx.Done()
	}()

	return DefaultConfig

}

func GetConfig(c Config) *viper.Viper {
	if DefaultConfig == nil {
		switch strings.ToLower(c.source) {
		case "file":
			if c.path == "" || c.configType == "" {
				return nil
			}
			DefaultConfig = initFileConfig(c.path, c.configType)
			break
		case "consul":
			if c.address == "" || c.path == "" || c.configType == "" {
				return nil
			}
			DefaultConfig = initConsulConfig(c.address, c.path, c.configType)
			break
		default:
			return nil
		}
	}
	return DefaultConfig
}
