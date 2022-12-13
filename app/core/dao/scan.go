package dao

import (
	"context"
	"fmt"
	"human/app/core/model"
	"human/library/log"
)

const (
	TSCAN = "t_scan"
)

var (
	_insertscanSQL = fmt.Sprintf("INSERT INTO %s(user_id,scan_id,status)VALUES(?,?,?)", TSCAN)
)

// data list
func (d *Dao) GetScan(c context.Context) (list []*model.TScan, err error) {
	list = make([]*model.TScan, 0)

	sql := fmt.Sprintf("SELECT Id,user_id,scan_id from %s where status = 1 or status=3", TSCAN)

	rows, err := d.db.Query(c, sql)
	if err != nil {
		log.Error("d.db.Query(%s) error(%v)", sql, err)
		return list, err
	}
	for rows.Next() {
		tmp := new(model.TScan)
		err = rows.Scan(&tmp.Id, &tmp.UserId, &tmp.ScanId)
		if err != nil {
			log.Error("rows.Scan error(%v)", err)
			continue
		}

		bearlist, _ := d.GetUserBearer(c, tmp.UserId)
		if len(bearlist) == 0 {
			continue
		}
		tmp.Bearer = bearlist[0].Bearer
		list = append(list, tmp)
	}

	return list, nil
}

// data list
func (d *Dao) GetUserScan(c context.Context, user_id int64) (list []*model.TScan, err error) {
	list = make([]*model.TScan, 0)

	sql := fmt.Sprintf("SELECT  scan_id,result_model_fbx,result_model_pre,status  from %s where  user_id=%d", TSCAN, user_id)

	rows, err := d.db.Query(c, sql)
	if err != nil {
		log.Error("d.db.Query(%s) error(%v)", sql, err)
		return list, err
	}
	for rows.Next() {
		tmp := new(model.TScan)
		err = rows.Scan(&tmp.ScanId, &tmp.ModelFbx,
			&tmp.ModelPre, &tmp.Status)
		if err != nil {
			log.Error("rows.Scan error(%v)", err)
			continue
		}

		list = append(list, tmp)
	}

	return list, nil
}

func (d *Dao) AddScan(c context.Context, scan *model.TScan) (lstid int64, err error) {
	res, err := d.db.Exec(c, _insertscanSQL, scan.UserId, scan.ScanId, scan.Status)
	if err != nil {
		log.Error("Register error(%v) : %v", err, scan)
		return
	}
	lstid, err = res.LastInsertId()
	return
}

func (d *Dao) UpdateScan(c context.Context, fields map[string]string, where map[string]string) (RowsAffected int64, err error) {
	sql := fmt.Sprintf("update %s set", TSCAN)
	key := ""
	value := make([]string, 0)

	for k, v := range fields {
		if key != "" {
			key += ","
		}
		key += " " + k + "=?"
		value = append(value, v)
	}
	sql += key
	w := ""
	for k, v := range where {
		if w == "" {
			w += " where "
		} else {
			w += " and "
		}
		w += k + "='" + v + "'"
	}
	sql += w

	if len(value) == 1 {
		res, _ := d.db.Exec(c, sql, value[0])
		RowsAffected, err = res.RowsAffected()
	} else if len(value) == 2 {
		res, _ := d.db.Exec(c, sql, value[0], value[1])
		RowsAffected, err = res.RowsAffected()
	} else if len(value) == 3 {
		res, _ := d.db.Exec(c, sql, value[0], value[1], value[2])
		RowsAffected, err = res.RowsAffected()
	} else if len(value) == 4 {
		res, _ := d.db.Exec(c, sql, value[0], value[1], value[2], value[3])
		RowsAffected, err = res.RowsAffected()
	}
	if err != nil {
		log.Error("Register error(%v)", err)
		return
	}

	return
}
