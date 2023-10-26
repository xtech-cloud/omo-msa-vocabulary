package nosql

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Record struct {
	UID         primitive.ObjectID `bson:"_id"`
	ID          uint64             `json:"id" bson:"id"`
	CreatedTime time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedTime time.Time          `json:"updatedAt" bson:"updatedAt"`
	DeleteTime  time.Time          `json:"deleteAt" bson:"deleteAt"`
	Created     int64              `json:"created" bson:"created"`
	Updated     int64              `json:"updated" bson:"updated"`
	Deleted     int64              `json:"deleted" bson:"deleted"`
	Creator     string             `json:"creator" bson:"creator"`
	Operator    string             `json:"operator" bson:"operator"`

	Option uint8  `json:"option" bson:"option"`
	From   string `json:"from" json:"from"`
	To     string `json:"to" bson:"to"`
	Entity string `json:"entity" bson:"entity"`
	Remark string `json:"remark" bson:"remark"`
}

func CreateRecord(info *Record) error {
	_, err := insertOne(TableRecord, info)
	if err != nil {
		return err
	}
	return nil
}

func GetRecordNextID() uint64 {
	num, _ := getSequenceNext(TableRecord)
	return num
}

func GetRecord(uid string) (*Record, error) {
	result, err := findOne(TableRecord, uid)
	if err != nil {
		return nil, err
	}
	model := new(Record)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetRecords(entity string) ([]*Record, error) {
	var items = make([]*Record, 0, 20)
	filter := bson.M{"entity": entity}
	cursor, err1 := findMany(TableRecord, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Record)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetRecordsBy(entity, to string, tp uint8) ([]*Record, error) {
	var items = make([]*Record, 0, 20)
	filter := bson.M{"entity": entity, "option": tp, "to": to}
	cursor, err1 := findMany(TableRecord, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Record)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetRecordsByRelate(to string, tp uint8) ([]*Record, error) {
	var items = make([]*Record, 0, 20)
	filter := bson.M{"option": tp, "to": to}
	cursor, err1 := findMany(TableRecord, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Record)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}
