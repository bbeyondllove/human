package conf

import (
	"flag"
	"human/library/cache/redis"
	"human/library/database/sql"
	"human/library/log"
	xhttp "human/library/net/http"
	xtime "human/library/time"

	"github.com/BurntSushi/toml"
)

var (
	httpPort int
	ConfPath string
	Conf     = &Config{}
)

type Config struct {
	// App
	App *App
	//Jwt
	Auth *Auth
	// Log
	Log *log.Config
	// DB
	MySQL *MySQL
	// Redis
	Redis *redis.Config
	// HttpClient
	HttpClient map[string]*xhttp.HttpClient
}

type App struct {
	HttpPort     int
	RunMode      string
	ReadTimeout  xtime.Duration
	WriteTimeout xtime.Duration
}

type Auth struct {
	SignKey    string
	ExpireTime int
	MaxTimeOut int
}

type MySQL struct {
	ConnInfo *sql.Config
}

func getConf(filePath string) {
	flag.IntVar(&httpPort, "http.port", -1, "http port")
	ConfPath = filePath
	flag.StringVar(&ConfPath, "conf", filePath, "config path")
}

func Init(filePath string) (err error) {
	getConf(filePath)
	_, err = toml.DecodeFile(ConfPath+"/application.toml", &Conf)
	return
}
