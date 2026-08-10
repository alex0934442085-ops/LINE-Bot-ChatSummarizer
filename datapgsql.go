package main

import (
	"log"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type PGSqlDB struct {
	Db *pg.DB
}

func (mdb *PGSqlDB) ReadGroupInfo(roomID string) GroupData {

	pgsql := &DBStorage{
		RoomID: roomID,
	}

	if ret, err := pgsql.Get(mdb); err == nil {
		return ret.Dataset
	} else {
		log.Println("DB read err:", err)
	}

	return GroupData{}
}

func (mdb *PGSqlDB) AppendGroupInfo(roomID string, m MsgDetail) {

	u := mdb.ReadGroupInfo(roomID)
	u = append(u, m)

	pgsql := &DBStorage{
		RoomID:  roomID,
		Dataset: u,
	}

	// 如果資料不存在，就新增
	if _, err := pgsql.Get(mdb); err != nil {

		pgsql.Add(mdb)
		return
	}

	// 如果已存在，就更新
	if err := pgsql.Update(mdb); err != nil {
		log.Println("DB update err:", err)
	}
}

func (mdb *PGSqlDB) GetLastSummaryTime(roomID string) time.Time {

	state := &SummaryState{
		RoomID: roomID,
	}

	if err := state.Get(mdb); err != nil {
		log.Println("Summary state read err:", err)
		return time.Time{}
	}

	return state.LastSummaryTime
}

func (mdb *PGSqlDB) SetLastSummaryTime(roomID string, t time.Time) {

	state := &SummaryState{
		RoomID:          roomID,
		LastSummaryTime: t,
	}

	// 先嘗試取得現有資料
	existing := &SummaryState{
		RoomID: roomID,
	}

	if err := existing.Get(mdb); err != nil {

		// 不存在就新增
		state.Add(mdb)
		return
	}

	// 已存在就更新
	if err := state.Update(mdb); err != nil {
		log.Println("Summary state update err:", err)
	}
}

func NewPGSql(url string) *PGSqlDB {

	options, err := pg.ParseURL(url)

	if err != nil {
		panic(err)
	}

	db := pg.Connect(options)

	err = createSchema(db)

	if err != nil {
		panic(err)
	}

	return &PGSqlDB{
		Db: db,
	}
}

func createSchema(db *pg.DB) error {

	// 原本的資料表
	models := []interface{}{
		(*MemStorage)(nil),
		(*DBStorage)(nil),
		(*SummaryState)(nil),
	}

	for _, model := range models {

		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// ============================================================
// DBStorage
// ============================================================

type DBStorage struct {
	Id      int64     `bson:"_id"`
	RoomID  string    `json:"roomid" bson:"roomid"`
	Dataset GroupData `json:"dataset" bson:"dataset"`
}

func (u *DBStorage) Add(conn *PGSqlDB) {

	_, err := conn.Db.Model(u).Insert()

	if err != nil {
		log.Println("DB insert err:", err)
	}
}

func (u *DBStorage) Get(conn *PGSqlDB) (result *DBStorage, err error) {

	log.Println("*** Get dataset roomID =", u.RoomID)

	data := DBStorage{}

	err = conn.Db.Model(&data).
		Where("roomid = ?", u.RoomID).
		Select()

	if err != nil {
		log.Println("DB get err:", err)
		return nil, err
	}

	log.Println("DB result =", data)

	return &data, nil
}

func (u *DBStorage) Update(conn *PGSqlDB) (err error) {

	log.Println("*** Update DB group data =", u)

	_, err = conn.Db.Model(u).
		Set("dataset = ?", u.Dataset).
		Where("roomid = ?", u.RoomID).
		Update()

	if err != nil {
		log.Println("DB update err:", err)
	}

	return err
}

// ============================================================
// SummaryState
//
// 每個群組獨立記錄「上一次統整完成時間」
// ============================================================

type SummaryState struct {
	Id              int64     `pg:",pk"`
	RoomID          string    `pg:",unique"`
	LastSummaryTime time.Time
}

func (u *SummaryState) Add(conn *PGSqlDB) {

	_, err := conn.Db.Model(u).Insert()

	if err != nil {
		log.Println("Summary state insert err:", err)
	}
}

func (u *SummaryState) Get(conn *PGSqlDB) error {

	return conn.Db.Model(u).
		Where("room_id = ?", u.RoomID).
		Select()
}

func (u *SummaryState) Update(conn *PGSqlDB) error {

	_, err := conn.Db.Model(u).
		Set("last_summary_time = ?", u.LastSummaryTime).
		Where("room_id = ?", u.RoomID).
		Update()

	return err
}
