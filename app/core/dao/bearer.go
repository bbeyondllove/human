package dao

import (
	"context"
	"fmt"
	"human/app/core/model"
	"human/library/log"
)

const (
	TBEARER = "t_bearer"
)

// data list
func (d *Dao) GetBearer(c context.Context) ([]*model.TBearer, error) {
	sql := fmt.Sprintf("SELECT bearer,user_id  from %s where user_id=0 ", TBEARER)
	list := make([]*model.TBearer, 0)

	rows, err := d.db.Query(c, sql)
	if err != nil {
		log.Error("d.db.Query(%s) error(%v)", sql, err)
		return list, err
	}
	for rows.Next() {
		tmp := new(model.TBearer)
		err = rows.Scan(&tmp.Bearer, &tmp.UserId)
		if err != nil {
			log.Error("rows.Scan error(%v)", err)
			continue
		}
		list = append(list, tmp)
	}
	return list, nil
}

// data list
func (d *Dao) GetUserBearer(c context.Context, user_id int64) ([]*model.TBearer, error) {
	sql := fmt.Sprintf("SELECT bearer,user_id  from %s where user_id=%d", TBEARER, user_id)
	list := make([]*model.TBearer, 0)

	rows, err := d.db.Query(c, sql)
	if err != nil {
		log.Error("d.db.Query(%s) error(%v)", sql, err)
		return list, err
	}
	for rows.Next() {
		tmp := new(model.TBearer)
		err = rows.Scan(&tmp.Bearer, &tmp.UserId)
		if err != nil {
			log.Error("rows.Scan error(%v)", err)
			continue
		}
		list = append(list, tmp)
	}
	return list, nil
}

// Update
func (d *Dao) UpdateBearer(c context.Context, bearer *model.TBearer) (RowsAffected int64, err error) {

	sql := fmt.Sprintf("update %s set user_id=? where bearer=?", TBEARER)
	res, err := d.db.Exec(c, sql, bearer.UserId, bearer.Bearer)
	if err != nil {
		log.Error("Register error(%v) : %v", err, bearer)
		return
	}
	RowsAffected, err = res.RowsAffected()
	return
}
