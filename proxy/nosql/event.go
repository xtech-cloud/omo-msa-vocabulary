package nosql

import (
	"context"
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
	Creator     string             `json:"creator" bson:"creator"`
	Operator    string             `json:"operator" bson:"operator"`

	Type        uint8                    `json:"type" bson:"type"`
	Access  	uint8 					 `json:"access" bson:"access"`
	Entity      string                   `json:"entity" bson:"entity"`
	Name        string                   `json:"name" bson:"name"`
	Description string                   `json:"desc" bson:"desc"`
	Cover       string                   `json:"cover" bson:"cover"`
	Quote       string `json:"quote" bson:"quote"`
	Date        proxy.DateInfo           `json:"date" bson:"date"`
	Place       proxy.PlaceInfo          `json:"place" bson:"place"`
	Tags        []string                 `json:"tags" bson:"tags"`
	Assets      []string                 `json:"assets" bson:"assets"`
	Relations   []proxy.RelationCaseInfo `json:"relations" bson:"relations"`
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

func GetRelationCaseNextID() uint64 {
	num, _ := getSequenceNext(TableRelationCase)
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

func GetEventsByTypeAndAccess(entity string,tp, access uint8) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	def := new(time.Time)
	filter := bson.M{"entity": entity, "type":tp, "access":access, "deleteAt": def}
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

func GetEventsByAccess(entity string, access uint8) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	def := new(time.Time)
	filter := bson.M{"entity": entity, "access":access, "deleteAt": def}
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

func GetEventsByQuote(entity, quote string) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	def := new(time.Time)
	filter := bson.M{"entity": entity, "quote":quote, "deleteAt": def}
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

func GetEventsByQuote2(quote string) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	def := new(time.Time)
	filter := bson.M{"quote":quote, "deleteAt": def}
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

func GetEventsByType(entity string, tp uint8) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	def := new(time.Time)
	filter := bson.M{"entity": entity, "type":tp, "deleteAt": def}
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

func GetEventsByTypeQuote(entity, quote string, tp uint8) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	def := new(time.Time)
	filter := bson.M{"entity": entity, "quote":quote, "type":tp, "deleteAt": def}
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

func UpdateEventBase(uid, name, desc, operator string, access uint8, date proxy.DateInfo, place proxy.PlaceInfo, assets []string) error {
	msg := bson.M{"name": name, "desc": desc, "assets": assets, "date": date, "access":access,
		"place": place, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventAccess(uid, operator string, access uint8) error {
	msg := bson.M{"access": access, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventQuote(uid, quote, operator string) error {
	msg := bson.M{"quote": quote, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventInfo(uid, name, desc, operator string) error {
	msg := bson.M{"name": name, "desc": desc, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventTags(uid, operator string, list []string) error {
	msg := bson.M{"tags": list, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventAssets(uid, operator string, list []string) error {
	msg := bson.M{"assets": list, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventCover(uid, operator, cover string) error {
	msg := bson.M{"cover": cover, "operator": operator, "updatedAt": time.Now()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func AppendEventAsset(uid string, asset string) error {
	msg := bson.M{"assets": asset}
	_, err := appendElement(TableEvent, uid, msg)
	return err
}

func SubtractEventAsset(uid string, asset string) error {
	msg := bson.M{"assets": asset}
	_, err := removeElement(TableEvent, uid, msg)
	return err
}

func AppendEventRelation(uid string, relation *proxy.RelationCaseInfo) error {
	msg := bson.M{"relations": relation}
	_, err := appendElement(TableEvent, uid, msg)
	return err
}

func SubtractEventRelation(uid string, relation string) error {
	msg := bson.M{"relations": bson.M{"uid": relation}}
	_, err := removeElement(TableEvent, uid, msg)
	return err
}
