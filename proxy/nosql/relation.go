package nosql

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Relation struct {
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

	Name   string `json:"name" bson:"name"`
	Remark string `json:"remark" bson:"remark"`
	Custom bool   `json:"custom" bson:"custom"`
	Type   uint8  `json:"type" bson:"type"`
	Parent string `json:"parent" bson:"parent"`
}

func CreateRelation(info *Relation) error {
	_, err := insertOne(TableRelation, info)
	if err != nil {
		return err
	}
	return nil
}

func GetRelationNextID() uint64 {
	num, _ := getSequenceNext(TableRelation)
	return num
}

func GetRelationsByParent(parent string) ([]*Relation, error) {
	var items = make([]*Relation, 0, 20)
	filter := bson.M{"parent": parent, TimeDeleted: 0}
	cursor, err1 := findMany(TableRelation, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Relation)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetRelation(uid string) (*Relation, error) {
	result, err := findOne(TableRelation, uid)
	if err != nil {
		return nil, err
	}
	model := new(Relation)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetRelationByName(name string) (*Relation, error) {
	msg := bson.M{"name": name, TimeDeleted: 0}
	result, err := findOneBy(TableRelation, msg)
	if err != nil {
		return nil, err
	}
	model := new(Relation)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetTopRelations() ([]*Relation, error) {
	var items = make([]*Relation, 0, 100)
	filter := bson.M{"parent": "", TimeDeleted: 0}
	cursor, err1 := findMany(TableRelation, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Relation)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func RemoveRelation(uid, operator string) error {
	_, err := removeOne(TableRelation, uid, operator)
	return err
}

func UpdateRelationBase(uid, name, desc, operator string, custom bool, kind uint8) error {
	msg := bson.M{"name": name, "remark": desc, "custom": custom, "type": kind, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableRelation, uid, msg)
	return err
}
