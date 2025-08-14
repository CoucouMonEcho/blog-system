package infrastructure

import (
	"sync"
	"time"
)

// PVAggregator 简易内存聚合器：仅用于演示与占位
type PVAggregator struct {
	mu         sync.RWMutex
	byMinute   map[int64]int64              // ts(分钟对齐) -> pv
	usersByDay map[int64]map[int64]struct{} // day(yyyyMMdd) -> user_id set
}

func NewPVAggregator() *PVAggregator {
	return &PVAggregator{
		byMinute:   make(map[int64]int64),
		usersByDay: make(map[int64]map[int64]struct{}),
	}
}

func bucketMinute(ts time.Time) int64 {
	// 对齐到分钟
	return ts.Unix() / 60 * 60
}

func dayKey(ts time.Time) int64 {
	y, m, d := ts.Date()
	return int64(y*10000 + int(m)*100 + d)
}

// RecordPV 记录 PV/UV（当 userID>0 时记为 UV）
func (a *PVAggregator) RecordPV(now time.Time, userID *int64) {
	min := bucketMinute(now)
	day := dayKey(now)
	a.mu.Lock()
	a.byMinute[min]++
	if userID != nil && *userID > 0 {
		set, ok := a.usersByDay[day]
		if !ok {
			set = make(map[int64]struct{})
			a.usersByDay[day] = set
		}
		set[*userID] = struct{}{}
	}
	a.mu.Unlock()
}

// Overview 返回今日 PV/UV 与近5分钟在线人数（近5分钟内有行为的用户数）
func (a *PVAggregator) Overview(now time.Time) (pvToday int64, uvToday int64, onlineUsers int64) {
	day := dayKey(now)
	from := now.Add(-5 * time.Minute)
	minFrom := bucketMinute(from)
	minTo := bucketMinute(now)
	a.mu.RLock()
	for m := minFrom; m <= minTo; m += 60 {
		pvToday += a.byMinute[m]
	}
	if set, ok := a.usersByDay[day]; ok {
		uvToday = int64(len(set))
	}
	a.mu.RUnlock()
	// 近5分钟在线用户（估算）：使用 uvToday 作为近似占位
	onlineUsers = uvToday
	return
}

// PVTimeSeries 简单按照 interval 聚合（支持 5m/1h/1d）
func (a *PVAggregator) PVTimeSeries(from, to time.Time, interval time.Duration) []struct {
	Ts    int64
	Value int64
} {
	res := make([]struct {
		Ts    int64
		Value int64
	}, 0)
	if from.After(to) {
		return res
	}
	// 归一化到分钟
	start := time.Unix(bucketMinute(from), 0)
	end := time.Unix(bucketMinute(to), 0)
	a.mu.RLock()
	for cur := start; !cur.After(end); cur = cur.Add(interval) {
		var val int64
		// 在区间 [cur, cur+interval) 求和
		for m := cur; m.Before(cur.Add(interval)); m = m.Add(time.Minute) {
			val += a.byMinute[bucketMinute(m)]
		}
		res = append(res, struct {
			Ts    int64
			Value int64
		}{Ts: cur.Unix(), Value: val})
	}
	a.mu.RUnlock()
	return res
}
