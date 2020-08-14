package nosql

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy"
	"time"
)

type Entity struct {
	UID         primitive.ObjectID    `bson:"_id"`
	ID          uint64                `json:"id" bson:"id"`
	CreatedTime time.Time             `json:"createdAt" bson:"createdAt"`
	UpdatedTime time.Time             `json:"updatedAt" bson:"updatedAt"`
	DeleteTime  time.Time             `json:"deleteAt" bson:"deleteAt"`
	Creator     string                `json:"creator" bson:"creator"`
	Operator    string                `json:"operator" bson:"operator"`

	Name        string                `json:"name" bson:"name"`
	Description string                `json:"desc" bson:"desc"`
	Cover       string                `json:"cover" bson:"cover"`
	Concept     string                `json:"concept" bson:"concept"`
	Status      uint8                 `json:"status" bson:"status"`
	Scene       string                `json:"scene" bson:"scene"` // 所属场景
	Add         string                `json:"add" bson:"add"`
	Synonyms    []string              `json:"synonyms" bson:"synonyms"`
	Tags        []string              `json:"tags" bson:"tags"`
	Properties  []*proxy.PropertyInfo `json:"props" bson:"props"`
}

func CreateEntity(info interface{}, table string) error {
	_, err := insertOne(table, info)
	if err != nil {
		return err
	}
	return nil
}

func GetEntities(table string) ([]*Entity, error) {
	cursor, err1 := findAll(table, 0)
	if err1 != nil {
		return nil, err1
	}
	var items = make([]*Entity, 0, 100)
	for cursor.Next(context.Background()) {
		var node = new(Entity)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetEntityNextID(table string) uint64 {
	num, _ := getSequenceNext(table)
	return num
}

func GetNodeNextID() uint64 {
	num, _ := getSequenceNext("nodes")
	return num
}

func GetLinkNextID() uint64 {
	num, _ := getSequenceNext("links")
	return num
}

func RemoveEntity(table string, uid string, operator string) error {
	_, err := removeOne(table, uid, operator)
	return err
}

func GetEntity(table string, uid string) (*Entity, error) {
	result, err := findOne(table, uid)
	if err != nil {
		return nil, err
	}
	model := new(Entity)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetEntityByName(table string, name string) (*Entity, error) {
	msg := bson.M{"name": name }
	result, err := findOneBy(table, msg)
	if err != nil {
		return nil, err
	}
	model := new(Entity)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func UpdateEntityBase(table, uid, name, remark, add, concept, operator string) error {
	msg := bson.M{"name": name, "remark": remark, "add": add, "concept": concept, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntityStatus(table string, uid string, state uint8, operator string) error {
	msg := bson.M{"status": state, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntityCover(table string, uid string, cover string, operator string) error {
	msg := bson.M{"cover": cover, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntityTags(table string, uid string, operator string, tags []string) error {
	msg := bson.M{"tags": tags, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntityAdd(table string, uid string, add string, operator string) error {
	msg := bson.M{"add": add, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntitySynonyms(table string, uid string, operator string, synonyms []string) error {
	msg := bson.M{"synonyms": synonyms, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntityProperties(table, uid string, operator string, array []*proxy.PropertyInfo) error {
	msg := bson.M{"props": array, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func AppendEntityProperty(table string, uid string, prop proxy.PropertyInfo) error {
	if len(uid) < 1 {
		return errors.New("the uid is empty")
	}
	msg := bson.M{"props": prop}
	_, err := appendElement(table, uid, msg)
	return err
}

func SubtractEntityProperty(table string, uid string, key string) error {
	if len(uid) < 1 {
		return errors.New("the uid is empty")
	}
	msg := bson.M{"props": bson.M{"key": key}}
	_, err := removeElement(table, uid, msg)
	return err
}
