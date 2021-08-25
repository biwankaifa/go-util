package ipip

type Ip struct {
	Address string `form:"address" json:"address"`
	CityID  int    `form:"city_id" json:"city_id"`
}

func Get(ip string) (ipInfo Ip) {
	//db, err := ipdb.NewCity("app/util/ipip/ipipfree.ipdb")
	//if err != nil {
	//	return
	//}
	//info, err := db.FindInfo(ip, "CN")
	//if err != nil {
	//	return
	//}
	ipInfo.CityID = 1
	ipInfo.Address = ip
	return
}
