package nosql

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

/**
概念定义的属性
*/
type Attribute struct {
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

	Kind   uint8  `json:"type" bson:"type"`
	Key    string `json:"key" bson:"key"`
	Name   string `json:"name" bson:"name"`
	Remark string `json:"remark" bson:"remark"`
	Begin  string `json:"begin" bson:"begin"`
	End    string `json:"end" bson:"end"`
}

func CreateAttribute(info *Attribute) error {
	_, err := insertOne(TableAttribute, info)
	if err != nil {
		return err
	}
	return nil
}

func GetAttributeNextID() uint64 {
	num, _ := getSequenceNext(TableAttribute)
	return num
}

func GetAttribute(uid string) (*Attribute, error) {
	result, err := findOne(TableAttribute, uid)
	if err != nil {
		return nil, err
	}
	model := new(Attribute)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetAttributeByName(name string) (*Attribute, error) {
	msg := bson.M{"name": name, TimeDeleted: 0}
	result, err := findOneBy(TableAttribute, msg)
	if err != nil {
		return nil, err
	}
	model := new(Attribute)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetAttributeByKey(key string) (*Attribute, error) {
	msg := bson.M{"key": key, TimeDeleted: 0}
	result, err := findOneBy(TableAttribute, msg)
	if err != nil {
		return nil, err
	}
	model := new(Attribute)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetAllAttributes() ([]*Attribute, error) {
	var items = make([]*Attribute, 0, 100)
	cursor, err1 := findAllEnable(TableAttribute, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Attribute)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func RemoveAttribute(uid, operator string) error {
	_, err := removeOne(TableAttribute, uid, operator)
	return err
}

func UpdateAttributeBase(uid, name, desc, begin, end, operator string, kind uint8) error {
	msg := bson.M{"name": name, "remark": desc, "type": kind, "begin": begin, "end": end, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableAttribute, uid, msg)
	return err
}

func UpdateAttributeKey(uid, key, operator string) error {
	msg := bson.M{"key": key, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableAttribute, uid, msg)
	return err
}
