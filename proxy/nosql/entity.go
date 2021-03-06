package nosql

import (
	"context"
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
	Summary     string                `json:"summary" bson:"summary"`
	Cover       string                `json:"cover" bson:"cover"`
	Concept     string                `json:"concept" bson:"concept"`
	Status      uint8                 `json:"status" bson:"status"`
	Scene       string                `json:"scene" bson:"scene"` // 所属场景
	Add         string                `json:"add" bson:"add"`
	Mark        string                `json:"mark" bson:"mark"`
	Quote       string                `json:"quote" bson:"quote"`
	Synonyms    []string              `json:"synonyms" bson:"synonyms"`
	Tags        []string              `json:"tags" bson:"tags"`
	Properties  []*proxy.PropertyInfo `json:"props" bson:"props"`
	Events      []*proxy.EventBrief 	`json:"events" bson:"events"`
	Relations   []*proxy.RelationCaseInfo `json:"relations" bson:"relations"`
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

func RemoveEntity(table, uid string, operator string) error {
	_, err := removeOne(table, uid, operator)
	return err
}

func GetEntity(table, uid string) (*Entity, error) {
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

func GetEntityByName(table, name, add string) (*Entity, error) {
	msg := bson.M{"name": name, "add": add, "deleteAt": new(time.Time)}
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

func GetEntityByMark(table, mark string) (*Entity, error) {
	msg := bson.M{"mark": mark}
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

func GetEntitiesByProp(table, key, value string) ([]*Entity, error) {
	msg := bson.M{"props": bson.M{"$elemMatch":bson.M{"key":key, "values": bson.M{"$elemMatch":bson.M{"name":value}}}}}
	cursor, err1 := findMany(table, msg, 0)
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

func GetEntitiesByOwnerAndStatus(table, owner string, st uint8) ([]*Entity, error) {
	msg := bson.M{"scene": owner, "status": st, "deleteAt": new(time.Time)}
	cursor, err1 := findMany(table, msg, 0)
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

func GetEntitiesByConcept(table, concept string) ([]*Entity, error) {
	msg := bson.M{"concept": concept, "deleteAt": new(time.Time)}
	cursor, err1 := findMany(table, msg, 0)
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

func GetEntitiesByName(table, name string) ([]*Entity, error) {
	msg := bson.M{"name": name, "deleteAt": new(time.Time)}
	cursor, err1 := findMany(table, msg, 0)
	if err1 != nil {
		return nil, err1
	}
	var items = make([]*Entity, 0, 10)
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

func GetEntitiesByOwner(table, owner string) ([]*Entity, error) {
	msg := bson.M{"scene": owner, "deleteAt": new(time.Time)}
	cursor, err1 := findMany(table, msg, 0)
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

func GetEntitiesByMatch(table, name string) ([]*Entity, error) {
	msg := bson.M{"name": bson.M{"$regex": name}, "deleteAt": new(time.Time)}
	cursor, err1 := findMany(table, msg, 0)
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

func GetEntitiesByStatus(table string, st uint8) ([]*Entity, error) {
	msg := bson.M{"status": st, "deleteAt": new(time.Time)}
	cursor, err1 := findMany(table, msg, 0)
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

func UpdateEntityBase(table, uid, name, remark, add, concept, quote, operator string) error {
	msg := bson.M{"name": name, "desc": remark, "add": add, "quote": quote, "concept": concept, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntityStatic(table, uid, operator string, tags []string, props []*proxy.PropertyInfo, events []*proxy.EventBrief, relations []*proxy.RelationCaseInfo) error {
	msg := bson.M{"operator": operator, "updatedAt": time.Now(),"tags": tags, "props": props, "events": events, "relations":relations}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntityStatus(table, uid string, state uint8, operator string) error {
	msg := bson.M{"status": state, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntityCover(table, uid string, cover string, operator string) error {
	msg := bson.M{"cover": cover, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntityTags(table, uid string, operator string, tags []string) error {
	msg := bson.M{"tags": tags, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntityAdd(table, uid string, add string, operator string) error {
	msg := bson.M{"add": add, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntitySynonyms(table, uid string, operator string, synonyms []string) error {
	msg := bson.M{"synonyms": synonyms, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func UpdateEntityProperties(table, uid string, operator string, array []*proxy.PropertyInfo) error {
	msg := bson.M{"props": array, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(table, uid, msg)
	return err
}

func AppendEntityProperty(table, uid string, prop proxy.PropertyInfo) error {
	msg := bson.M{"props": prop}
	_, err := appendElement(table, uid, msg)
	return err
}

func SubtractEntityProperty(table, uid string, key string) error {
	msg := bson.M{"props": bson.M{"key": key}}
	_, err := removeElement(table, uid, msg)
	return err
}
