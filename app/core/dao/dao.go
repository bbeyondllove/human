package dao

import (
	"context"
	"human/app/core/conf"
	"human/library/cache/redis"
	xsql "human/library/database/sql"
	"human/library/log"
	xhttp "human/library/net/http"
)

// Dao struct
type Dao struct {
	Conf *conf.Config
	// mysql
	db *xsql.DB
	// redis
	redis *redis.Pool
	// httpClient
	In3dClient *xhttp.Client
	// httpClient
	AsposeClient *xhttp.Client
}

// New init
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		Conf:         c,
		db:           xsql.NewMySQL(c.MySQL.ConnInfo),
		redis:        redis.NewPool(c.Redis),
		In3dClient:   xhttp.NewClient(c.HttpClient["in3d"]),
		AsposeClient: xhttp.NewClient(c.HttpClient["aspose"]),
	}

	return
}

// BeginTran begin transaction
func (d *Dao) BeginTran(ctx context.Context) (tx *xsql.Tx, err error) {
	if tx, err = d.db.Begin(ctx); err != nil {
		log.Error("BeginTran d.arcDB.Begin error(%v)", err)
	}
	return
}

// Ping ping the resource.
func (d *Dao) Ping(ctx context.Context) (err error) {
	// TODO
	return
}

// Close close the resource.
func (d *Dao) Close() {
	d.db.Close()
}
