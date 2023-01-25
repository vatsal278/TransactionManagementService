package config

import (
	"database/sql"
	"fmt"
	"github.com/PereRohit/util/config"
	"github.com/vatsal278/TransactionManagementService/internal/model"
	"github.com/vatsal278/TransactionManagementService/internal/repo/authentication"
	"github.com/vatsal278/go-redis-cache"
	"time"
)

type Config struct {
	ServiceRouteVersion string              `json:"service_route_version"`
	ServerConfig        config.ServerConfig `json:"server_config"`
	// add custom config structs below for any internal services
	DataBase  DbCfg        `json:"db_svc"`
	SecretKey string       `json:"secret_key"`
	Cookie    CookieStruct `json:"cookie"`
	Cache     CacheCfg     `json:"cache"`
}

type SvcConfig struct {
	Cfg                 *Config
	ServiceRouteVersion string
	SvrCfg              config.ServerConfig
	// add internal services after init
	DbSvc  DbSvc
	JwtSvc JWTSvc
	Cacher CacherSvc
}
type DbSvc struct {
	Db *sql.DB
}
type DbCfg struct {
	Port      string `json:"dbPort"`
	Host      string `json:"dbHost"`
	Driver    string `json:"dbDriver"`
	User      string `json:"dbUser"`
	Pass      string `json:"dbPass"`
	DbName    string `json:"dbName"`
	TableName string `json:"tableName"`
}
type JWTSvc struct {
	JwtSvc authentication.JWTService
}
type MsgQueueCfg struct {
	SvcUrl                  string   `json:"service_url"`
	AllowedUrl              []string `json:"allowed_url"`
	UserAgent               string   `json:"user_agent"`
	UrlCheck                bool     `json:"url_check_flag"`
	NewAccountChannel       string   `json:"new_account_channel"`
	ActivatedAccountChannel string   `json:"account_activation_channel"`
	Key                     string   `json:"private_key"`
}

type CookieStruct struct {
	Name      string        `json:"name"`
	Expiry    time.Duration `json:"-"`
	ExpiryStr string        `json:"expiry"`
	Path      string        `json:"path"`
}
type CacheCfg struct {
	Port     string `json:"port"`
	Host     string `json:"host"`
	Duration string `json:"duration"`
	Time     time.Duration
}
type CacherSvc struct {
	Cacher redis.Cacher
}

func Connect(cfg DbCfg, tableName string) *sql.DB {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True", cfg.User, cfg.Pass, cfg.Host, cfg.Port)
	db, err := sql.Open(cfg.Driver, connectionString)
	if err != nil {
		panic(err.Error())
	}
	dbString := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s ;", cfg.DbName)
	prepare, err := db.Prepare(dbString)
	if err != nil {
		panic(err.Error())
	}
	_, err = prepare.Exec()
	if err != nil {
		panic(err.Error())
	}
	db.Close()
	connectionString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True", cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.DbName)
	db, err = sql.Open(cfg.Driver, connectionString)
	if err != nil {
		panic(err.Error())
	}
	x := fmt.Sprintf("create table if not exists %s", tableName)
	_, err = db.Exec(x + model.Schema)
	if err != nil {
		panic(err.Error())
	}
	return db
}

func InitSvcConfig(cfg Config) *SvcConfig {
	// init required services and assign to the service struct fields
	dataBase := Connect(cfg.DataBase, cfg.DataBase.TableName)
	jwtSvc := authentication.JWTAuthService(cfg.SecretKey)
	cacher := redis.NewCacher(redis.Config{Addr: cfg.Cache.Host + ":" + cfg.Cache.Port})
	duration, err := time.ParseDuration(cfg.Cache.Duration)
	if err != nil {
		return nil
	}
	cfg.Cache.Time = duration
	return &SvcConfig{
		Cfg:                 &cfg,
		ServiceRouteVersion: cfg.ServiceRouteVersion,
		SvrCfg:              cfg.ServerConfig,
		DbSvc:               DbSvc{Db: dataBase},
		JwtSvc:              JWTSvc{JwtSvc: jwtSvc},
		Cacher:              CacherSvc{Cacher: cacher},
	}
}
