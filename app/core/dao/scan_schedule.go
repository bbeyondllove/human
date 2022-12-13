package dao

import (
	"context"
	"fmt"
	"human/app/core/model"
	"human/library/log"
	"time"
)

const (
	TSCHEDULE = "t_scan_schedule"
)

// data list
func (d *Dao) GetScheduler(c context.Context, scheduler *model.TScanScheduler) (list []*model.TScanScheduler, err error) {
	list = make([]*model.TScanScheduler, 0)
	if len(scheduler.Bearer) == 0 {
		return list, nil
	}

	sql := fmt.Sprintf("SELECT bearer,scan_id, count,update_time from %s where bearer = '%s' ", TSCHEDULE, scheduler.Bearer)

	rows, err := d.db.Query(c, sql)
	if err != nil {
		log.Error("d.db.Query(%s) error(%v)", sql, err)
		return list, err
	}
	for rows.Next() {
		tmp := new(model.TScanScheduler)
		err = rows.Scan(&tmp.Bearer, &tmp.ScanId, &tmp.Count, &tmp.UpdateTime)
		if err != nil {
			log.Error("rows.Scan error(%v)", err)
			continue
		}

		curtime := time.Now()
		if tmp.UpdateTime.Year() != curtime.Year() ||
			tmp.UpdateTime.Month() != curtime.Month() ||
			tmp.UpdateTime.Day() != curtime.Day() {
			continue
		}
		list = append(list, tmp)
	}
	return list, nil
}

// Update
func (d *Dao) UpdateScheduler(c context.Context, scheduler *model.TScanScheduler) (RowsAffected int64, err error) {

	//sql := fmt.Sprintf("INSERT INTO %s(bearer, scan_id)VALUES(?, ?) ON DUPLICATE KEY UPDATE scan_id = '%s'", TSCHEDULE, scheduler.ScanId)
	sql := fmt.Sprintf("INSERT INTO %s(bearer,user_id,scan_id)VALUES(?,?, ?) ", TSCHEDULE)
	res, err := d.db.Exec(c, sql, scheduler.Bearer, scheduler.UserId, scheduler.ScanId)
	if err != nil {
		log.Error("Register error(%v) : %v", err, scheduler)
		return
	}
	RowsAffected, err = res.RowsAffected()
	return
}
