package main

import "time"

type GroupDB interface {
	ReadGroupInfo(string) GroupData
	AppendGroupInfo(string, MsgDetail)

	// 取得這個群組上一次統整的時間
	GetLastSummaryTime(string) time.Time

	// 記錄這個群組這次統整完成的時間
	SetLastSummaryTime(string, time.Time)
}

type MsgDetail struct {
	MsgText  string
	UserName string
	Time     time.Time
}

type GroupData []MsgDetail
