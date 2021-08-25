package config

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go-util/env"
	"strconv"
	"strings"
	"time"
)

var config = new(Config)

type Config struct {
	App struct {
		Domain       string `toml:"domain"`
		Port         int    `toml:"port"`
		ReadTimeout  int    `toml:"readTimeout"`
		WriteTimeout int    `toml:"writeTimeout"`
	}

	Log struct {
		FilePath string `toml:"filePath"`
		FileName string `toml:"fileName"`
	}

	MySQL struct {
		Write struct {
			Addr string `toml:"addr"`
			User string `toml:"user"`
			Pass string `toml:"pass"`
			Name string `toml:"name"`
		} `toml:"read"`
		Base struct {
			MaxOpenConn     int           `toml:"maxOpenConn"`
			MaxIdleConn     int           `toml:"maxIdleConn"`
			ConnMaxLifeTime time.Duration `toml:"connMaxLifeTime"`
		} `toml:"base"`
	} `toml:"mysql"`

	Redis struct {
		Addr         string `toml:"addr"`
		Pass         string `toml:"pass"`
		Db           int    `toml:"db"`
		MaxRetries   int    `toml:"maxRetries"`
		PoolSize     int    `toml:"poolSize"`
		MinIdleConns int    `toml:"minIdleConns"`
	} `toml:"redis"`

	Rongcloud struct {
		AppKey    string `toml:"appKey"`
		AppSecret string `toml:"appSecret"`
	} `toml:"rongcloud"`

	Shanyan struct {
		Ios struct {
			AppID  string `toml:"appId"`
			AppKey string `toml:"appKey"`
		} `toml:"appId"`
		Android struct {
			AppID  string `toml:"appId"`
			AppKey string `toml:"appKey"`
		} `toml:"android"`
	} `toml:"shanyan"`

	QQ struct {
		AppID  string `toml:"appId"`
		AppKey string `toml:"appKey"`
	} `toml:"qq"`

	Wechat struct {
		AppID     string `toml:"appId"`
		AppSecret string `toml:"appSecret"`
	} `toml:"wechat"`

	Weibo struct {
		AppKey      string `toml:"appKey"`
		AppSecret   string `toml:"appSecret"`
		RedirectUri string `toml:"redirectUri"`
	} `toml:"weibo"`

	SMS struct {
		RongLianYun struct {
			AccountToken string `toml:"accountToken"`
			AccountSid   string `toml:"accountSid"`
			Domestic     struct {
				AppId      string `toml:"appId"`
				TemplateId int    `toml:"templateId"`
			} `toml:"domestic"`
			International struct {
				AppId      string `toml:"appId"`
				TemplateId int    `toml:"templateId"`
			} `toml:"international"`
		} `toml:"rongLianYun"`
	} `toml:"sms"`

	RabbitMQ struct {
		User string `toml:"user"`
		Pass string `toml:"pass"`
		Addr string `toml:"addr"`
	} `toml:"rabbitMQ"`

	Aliyun struct {
		Video struct {
			AccessKeyId     string `toml:"accessKeyId"`
			AccessKeySecret string `toml:"accessKeySecret"`
			Endpoint        string `toml:"endpoint"`
			TemplateGroupId string `toml:"templateGroupId"`
		} `toml:"video"`
		File struct {
			AccessKeyId     string `toml:"accessKeyId"`
			AccessKeySecret string `toml:"accessKeySecret"`
			Endpoint        string `toml:"endpoint"`
			RegionID        string `toml:"regionID"`
			BucketName      string `toml:"bucketName"`
		} `toml:"file"`
		Oss struct {
			AccessKeyId     string `toml:"accessKeyId"`
			AccessKeySecret string `toml:"accessKeySecret"`
			Endpoint        string `toml:"endpoint"`
			RegionID        string `toml:"regionID"`
			BucketName      string `toml:"bucketName"`
		} `toml:"oss"`
	} `toml:"aliyun"`

	Version struct {
		AndroidMin uint32 `toml:"androidMin"`
		IosMin     uint32 `toml:"iosMin"`
	} `toml:"version"`

	MongoDb struct {
		Host     string `toml:"host"`
		Port     int    `toml:"port"`
		Name     string `toml:"name"`
		Pass     string `toml:"pass"`
		DataBase string `toml:"database"`
		String   string `toml:"string"`
	} `toml:"mongoDB"`

	Payment struct {
		Alipay struct {
			AppID           string `toml:"appID"`
			AppPrivateKey   string `toml:"appPrivateKey"`
			AlipayPublicKey string `toml:"alipayPublicKey"`
			NotifyURL       string `toml:"notifyURL"`
			ReturnURL       string `toml:"returnURL"`
		} `toml:"payment"`
	} `toml:"alipay"`
}

func init() {
	viper.SetConfigName(env.Active().Value() + "_configs")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("config.toml not found")
		} else {
			panic(err)
		}
	}

	if err := viper.Unmarshal(config); err != nil {
		panic(err)
	}

	//创建一个信道等待关闭
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		//设置监听回调函数
		viper.OnConfigChange(func(e fsnotify.Event) {
			fmt.Printf("config is change :%s \n", e.String())
			if err := viper.Unmarshal(config); err != nil {
				panic(err)
			}
			cancel()
		})
		//开始监听
		viper.WatchConfig()
		//信道不会主动关闭，可以主动调用cancel关闭
		<-ctx.Done()
	}()

}

func Get() Config {
	return *config
}

func ProjectName() string {
	return "iFensi"
}

func RateLimit() int {
	return 10
}

// ProjectPort 运行端口号获取
func ProjectPort() (string string) {
	if config.App.Port != 0 {
		string = ":" + strconv.Itoa(config.App.Port)
	} else {
		string = ":80"
	}
	return
}

func ProjectLogFile() string {
	return fmt.Sprintf("./logs/%s-access.log", ProjectName())
}

// AppVersionMin App允许运行的最低版本号获取
func AppVersionMin(deviceOs string) uint32 {
	switch strings.ToLower(deviceOs) {
	case "android":
		return config.Version.AndroidMin
	case "ios":
		return config.Version.IosMin
	}
	return 0
}
