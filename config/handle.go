package config

import (
	"bytes"
	"context"
	"fmt"
	"github.com/biwankaifa/go-util/env"
	"github.com/fsnotify/fsnotify"
	"log"
	"time"

	consulApi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

type Config struct {
	// Source config获取来源方式, file 本地文件方式获取, consul 从consul的kv内获取
	Source string
	// Address consul下地址获取
	Address string
	// Path kv路径信息
	Path string
	// ConfigType kv数据格式
	ConfigType string
}

var defaultConfig *viper.Viper

func initConsulConfig(address, path, configType string) *viper.Viper {
	//consulAddress = "http://127.0.0.1:8500"
	//consulPath = "config"

	defaultConfig = viper.New()
	defaultConfig.SetConfigType(configType)

	consulClient, err := consulApi.NewClient(&consulApi.Config{Address: address})
	if err != nil {
		log.Fatalln("consul连接失败:", err)
	}

	kv, _, err := consulClient.KV().Get(path, nil)
	if err != nil {
		log.Fatalln("consul获取配置失败:", err)
	}

	err = defaultConfig.ReadConfig(bytes.NewBuffer(kv.Value))
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
			defaultConfig = hotconfig
		}
		err = w.Run(address)
		if err != nil {
			log.Fatalln("监听consul错误:", err)
		}
	}()

	return defaultConfig
}

func initFileConfig(active, path, configType string) *viper.Viper {
	defaultConfig = viper.New()
	defaultConfig.SetConfigName(active + "_configs")
	defaultConfig.SetConfigType(configType)
	defaultConfig.AddConfigPath("./" + path)

	if err := defaultConfig.ReadInConfig(); err != nil {
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
		defaultConfig.OnConfigChange(func(e fsnotify.Event) {
			fmt.Printf("config is change :%s \n", e.String())
			//if err := viper.Unmarshal(config); err != nil {
			//	panic(err)
			//}
			cancel()
		})
		//开始监听
		defaultConfig.WatchConfig()
		//信道不会主动关闭，可以主动调用cancel关闭
		<-ctx.Done()
	}()

	return defaultConfig

}

func (c Config) GetConfig() *viper.Viper {
	if defaultConfig == nil {
		switch c.ConfigType {
		case "file":
			defaultConfig = initConsulConfig(c.Address, c.Path, c.ConfigType)
			break
		case "consul":
			defaultConfig = initFileConfig(env.Active().Value(), c.Path, c.ConfigType)
			break
		}
	}
	return defaultConfig
}
