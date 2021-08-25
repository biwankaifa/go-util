package ipplus

import (
	"fmt"
	"github.com/ipplus360/awdb-golang/awdb-golang"
	"log"
	"net"
)

type Ip struct {
	Address string `form:"address" json:"address"`
	CityID  int    `form:"city_id" json:"city_id"`
}

func Get(ip string) (ipInfo Ip) {
	db, err := awdb.Open("app/util/ipplus/IP_trial_single_WGS84.awdb")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	ip1 := net.ParseIP("116.234.38.47")

	var record interface{}
	err = db.Lookup(ip1, &record)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", record)

	ipInfo.CityID = 1
	ipInfo.Address = ip
	return
}
