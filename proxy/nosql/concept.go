package nosql

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Concept struct {
	UID         primitive.ObjectID `bson:"_id"`
	ID          uint64             `json:"id" bson:"id"`
	CreatedTime time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedTime time.Time          `json:"updatedAt" bson:"updatedAt"`
	DeleteTime  time.Time          `json:"deleteAt" bson:"deleteAt"`
	Creator     string                `json:"creator" bson:"creator"`
	Operator    string                `json:"operator" bson:"operator"`

	Type   uint8 				 `json:"type" bson:"type"`
	Name   string                `json:"name" bson:"name"`
	Cover  string                `json:"cover" bson:"cover"`
	Remark string                `json:"remark" bson:"remark"`
	Table  string                `json:"table" bson:"table"`
	Parent string                `json:"parent" bson:"parent"`
	Scene  uint8 				 `json:"scene" bson:"scene"`
	Attributes  []string `json:"attributes" bson:"attributes"`
}

func CreateConcept(info *Concept) error {
	_, err := insertOne(TableConcept, info)
	if err != nil {
		return err
	}
	return nil
}

func GetConceptNextID() uint64 {
	num, _ := getSequenceNext(TableConcept)
	return num
}

func GetConcept(uid string) (*Concept, error) {
	result, err := findOne(TableConcept, uid)
	if err != nil {
		return nil, err
	}
	model := new(Concept)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetTopConcepts() ([]*Concept, error) {
	var items = make([]*Concept, 0, 20)
	def := new(time.Time)
	filter := bson.M{"parent": "", "deleteAt": def}
	cursor, err1 := findMany(TableConcept, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Concept)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetConceptsByParent(parent string) ([]*Concept, error) {
	var items = make([]*Concept, 0, 20)
	def := new(time.Time)
	filter := bson.M{"parent": parent, "deleteAt": def}
	cursor, err1 := findMany(TableConcept, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Concept)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func HadConceptByName(name string) (bool, error) {
	msg := bson.M{"name": name}
	return hadOne(TableConcept, msg)
}

func UpdateConceptBase(uid, name, desc, operator string, kind, scene uint8) error {
	msg := bson.M{"name": name, "remark": desc,"operator":operator,"type":kind, "scene":scene, "updatedAt": time.Now()}
	_, err := updateOne(TableConcept, uid, msg)
	return err
}

func UpdateConceptCover(uid string, icon string) error {
	msg := bson.M{"cover": icon, "updatedAt": time.Now()}
	_, err := updateOne(TableConcept, uid, msg)
	return err
}

func RemoveConcept(uid, operator string) error {
	_, err := removeOne(TableConcept, uid, operator)
	return err
}

func UpdateConceptAttributes(uid string, attrs []string) error {
	msg := bson.M{"attributes": attrs, "updatedAt": time.Now()}
	_, err := updateOne(TableConcept, uid, msg)
	return err
}

func AppendConceptAttribute(uid string, attr string) error {
	msg := bson.M{"attributes": attr}
	_, err := appendElement(TableConcept, uid, msg)
	return err
}

func SubtractConceptAttribute(uid string, attr string) error {
	msg := bson.M{"attributes": attr}
	_, err := removeElement(TableConcept, uid, msg)
	return err
}
