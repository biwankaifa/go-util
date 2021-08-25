package device_detector

import (
	"encoding/json"
	. "github.com/gamebtc/devicedetector"
	"github.com/gin-gonic/gin"
	"sync"
)

//单例模式
var deviceDetector *DeviceDetector
var once sync.Once

func init() {
	once.Do(func() {
		deviceDetector, _ = NewDeviceDetector("extend/device_detector")
	})
}

type DeviceInfos struct {
	Bot       bool   `json:"bot"`
	Os        VOS    `json:"os"`
	BrandName string `json:"brand_name"`
	Model     string `json:"model"`
}

type VOS struct {
	Name      string `yaml:"name" json:"name"`
	ShortName string `yaml:"short_name" json:"short_name"`
	Version   string `yaml:"version" json:"version"`
	Platform  string `yaml:"platform" json:"platform"`
}

func Get(c *gin.Context, u string) *DeviceInfos {
	// 获取用户设备信息
	wapDevices, _ := c.Cookie("g_w_device")
	if wapDevices == "" {
		wapDevice := DeviceInfos{}
		wapDeviceR := deviceDetector.Parse(u)
		wapDevice.Bot = wapDeviceR.IsBot()
		wapDevice.Model = wapDeviceR.GetModel()
		wapDevice.BrandName = wapDeviceR.GetBrandName()
		wapDevice.Os.Name = wapDeviceR.GetOs().Name
		wapDevice.Os.ShortName = wapDeviceR.GetOs().ShortName
		wapDevice.Os.Version = wapDeviceR.GetOs().Version
		//wapDevice.Os.Platform = wapDeviceR.GetClient().Platform

		// 转json
		result, _ := json.Marshal(wapDevice)
		c.SetCookie("g_w_device", string(result), 5, "/", "", false, true)
		return &wapDevice
	} else {
		wapDevice := DeviceInfos{}
		_ = json.Unmarshal([]byte(wapDevices), &wapDevice)
		return &wapDevice
	}
	//
}
