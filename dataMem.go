package main

import "time"

type MemStorage map[string]GroupData

type MemDB struct {
	db             MemStorage
	summaryRecords map[string]time.Time
}

func (mdb *MemDB) ReadGroupInfo(roomID string) GroupData {
	return mdb.db[roomID]
}

func (mdb *MemDB) AppendGroupInfo(roomID string, m MsgDetail) {
	mdb.db[roomID] = append(mdb.db[roomID], m)
}

func (mdb *MemDB) GetLastSummaryTime(roomID string) time.Time {
	return mdb.summaryRecords[roomID]
}

func (mdb *MemDB) SetLastSummaryTime(roomID string, t time.Time) {
	mdb.summaryRecords[roomID] = t
}

func NewMemDB() *MemDB {
	return &MemDB{
		db:             make(MemStorage),
		summaryRecords: make(map[string]time.Time),
	}
}
