package nosql

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"omo.msa.vocabulary/proxy"
	"time"
)

type Event struct {
	UID         primitive.ObjectID `bson:"_id"`
	ID          uint64             `json:"id" bson:"id"`
	CreatedTime time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedTime time.Time          `json:"updatedAt" bson:"updatedAt"`
	DeleteTime  time.Time          `json:"deleteAt" bson:"deleteAt"`
	Creator     string                `json:"creator" bson:"creator"`
	Operator    string                `json:"operator" bson:"operator"`

	Entity      string               `json:"entity" bson:"entity"`
	Name        string               `json:"name" bson:"name"`
	Description string               `json:"desc" bson:"desc"`
	Date        proxy.DateInfo       `json:"date" bson:"date"`
	Place       proxy.PlaceInfo      `json:"place" bson:"place"`
	Assets      []string             `json:"assets" bson:"assets"`
	Relations   []proxy.RelationInfo `json:"relations" bson:"relations"`
}

func CreateEvent(info *Event) error {
	_, err := insertOne(TableEvent, info)
	if err != nil {
		return err
	}
	return nil
}

func GetEventNextID() uint64 {
	num, _ := getSequenceNext(TableEvent)
	return num
}

func GetEvent(uid string) (*Event, error) {
	result, err := findOne(TableEvent, uid)
	if err != nil {
		return nil, err
	}
	model := new(Event)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetEventsByParent(parent string) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	def := new(time.Time)
	filter := bson.M{"entity": parent, "deleteAt": def}
	cursor, err1 := findMany(TableEvent, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Event)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func RemoveEvent(uid string, operator string) error {
	_, err := removeOne(TableEvent, uid, operator)
	return err
}

func UpdateEventBase(uid, name, desc, operator string, date proxy.DateInfo, place proxy.PlaceInfo) error {
	msg := bson.M{"name": name, "desc": desc, "date":date, "place":place, "operator":operator, "updatedAt": time.Now()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func AppendEventAsset(uid string, asset string) error {
	if len(uid) < 1 {
		return errors.New("the uid is empty")
	}
	msg := bson.M{"assets": asset}
	_, err := appendElement(TableConcept, uid, msg)
	return err
}

func SubtractEventAsset(uid string, asset string) error {
	if len(uid) < 1 {
		return errors.New("the uid is empty")
	}
	msg := bson.M{"assets": asset}
	_, err := removeElement(TableConcept, uid, msg)
	return err
}

func AppendEventRelation(uid string, relation proxy.RelationInfo) error {
	if len(uid) < 1 {
		return errors.New("the uid is empty")
	}
	msg := bson.M{"relations": relation}
	_, err := appendElement(TableConcept, uid, msg)
	return err
}

func SubtractEventRelation(uid string, relation string) error {
	if len(uid) < 1 {
		return errors.New("the uid is empty")
	}
	msg := bson.M{"relations": bson.M{ "uid": relation }}
	_, err := removeElement(TableConcept, uid, msg)
	return err
}
