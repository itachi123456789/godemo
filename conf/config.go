package conf

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	ListenAddr string
	Host       string
	CcpayHost  string
	Uid        string
	Secret     string
}

var (
	// AppConfig 应用程序配置
	AppConfig = &Config{}
)

func LoadConf() {
	// c := AppConfig
	r, err := os.Open("./conf/config.json")
	if err != nil {
		log.Fatalln(err)
	}

	decoder := json.NewDecoder(r)
	err = decoder.Decode(AppConfig)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%+v\n", AppConfig)
}
