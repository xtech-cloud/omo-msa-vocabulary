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
	Created     int64              `json:"created" bson:"created"`
	Updated     int64              `json:"updated" bson:"updated"`
	Deleted     int64              `json:"deleted" bson:"deleted"`
	Creator     string             `json:"creator" bson:"creator"`
	Operator    string             `json:"operator" bson:"operator"`

	Type        uint8                    `json:"type" bson:"type"`
	Access      uint8                    `json:"access" bson:"access"`
	Entity      string                   `json:"entity" bson:"entity"`
	Parent      string                   `json:"parent" bson:"parent"`
	Name        string                   `json:"name" bson:"name"`
	Description string                   `json:"desc" bson:"desc"`
	Cover       string                   `json:"cover" bson:"cover"`
	Quote       string                   `json:"quote" bson:"quote"`
	Owner       string                   `json:"owner" bson:"owner"`
	Date        proxy.DateInfo           `json:"date" bson:"date"`
	Place       proxy.PlaceInfo          `json:"place" bson:"place"`
	Tags        []string                 `json:"tags" bson:"tags"`
	Assets      []string                 `json:"assets" bson:"assets"`
	Targets     []string                 `json:"targets" bson:"targets"`
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

func GetEventByAsset(uid string) (*Event, error) {
	filter := bson.M{"assets": uid, TimeDeleted: 0}
	result, err := findOneBy(TableEvent, filter)
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

func GetEventByTarget(entity, target string) (*Event, error) {
	filter := bson.M{"entity": entity, "targets": target, TimeDeleted: 0}
	result, err := findOneBy(TableEvent, filter)
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

func GetEventCountByEntity(entity string) uint32 {
	filter := bson.M{"entity": entity, TimeDeleted: 0}
	count, err := getCountBy(TableEvent, filter)
	if err != nil {
		return 0
	}
	return uint32(count)
}

func GetEventCountByEntityTarget(entity, target string) uint32 {
	filter := bson.M{"entity": entity, "targets": target, TimeDeleted: 0}
	count, err := getCountBy(TableEvent, filter)
	if err != nil {
		return 0
	}
	return uint32(count)
}

func GetEventCountByType(entity string, tp uint8) uint32 {
	filter := bson.M{"entity": entity, "type": tp, TimeDeleted: 0}
	count, err := getCountBy(TableEvent, filter)
	if err != nil {
		return 0
	}
	return uint32(count)
}

func GetEventsByEntity(parent string) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	filter := bson.M{"entity": parent, TimeDeleted: 0}
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

func GetEventsByTypeAndAccess(entity string, tp, access uint8) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	filter := bson.M{"entity": entity, "type": tp, "access": access, TimeDeleted: 0}
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
	filter := bson.M{"entity": entity, "access": access, TimeDeleted: 0}
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
	filter := bson.M{"entity": entity, "quote": quote, TimeDeleted: 0}
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

func GetEventsByOwner(owner string) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	filter := bson.M{"owner": owner, TimeDeleted: 0}
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

func GetEventsAllByType(tp uint8) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	filter := bson.M{"type": tp, TimeDeleted: 0}
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
	filter := bson.M{"quote": quote, TimeDeleted: 0}
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

func GetEventsByDuration(quote string, from, to int64) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	filter := bson.M{"quote": quote, TimeCreated: bson.M{"$gt": from, "$lt": to}}
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

func GetEventsByRegex(quote, key, val string) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	filter := bson.M{"quote": quote, key: bson.M{"$regex": val}, TimeDeleted: 0}
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
	filter := bson.M{"entity": entity, "type": tp, TimeDeleted: 0}
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

func GetEventsByQuoteType(entity, quote string, tp uint8) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	filter := bson.M{"entity": entity, "quote": quote, "type": tp, TimeDeleted: 0}
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

func GetEventsByType2(tp uint8) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	filter := bson.M{"type": tp, TimeDeleted: 0}
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
	filter := bson.M{"entity": entity, "quote": quote, "type": tp, TimeDeleted: 0}
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

func GetEventsByTypeTarget(entity, target string, tp uint8) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	filter := bson.M{"entity": entity, "targets": target, "type": tp, TimeDeleted: 0}
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

func GetEventsByEntityTarget(entity, target string) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	filter := bson.M{"entity": entity, "targets": target, TimeDeleted: 0}
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

func GetEventsByTarget(target string) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	filter := bson.M{"targets": target, TimeDeleted: 0}
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

func GetEventsCountByOwnerTarget(owner, target string) uint32 {
	filter := bson.M{"owner": owner, "targets": target, TimeDeleted: 0}
	num, _ := getCountByFilter(TableEvent, filter)
	return uint32(num)
}

func GetEventsByOwnerTarget(owner, target string) ([]*Event, error) {
	var items = make([]*Event, 0, 20)
	filter := bson.M{"owner": owner, "targets": target, TimeDeleted: 0}
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
	msg := bson.M{"name": name, "desc": desc, "assets": assets, "date": date, "access": access,
		"place": place, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventAccess(uid, operator string, access uint8) error {
	msg := bson.M{"access": access, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventOwner(uid, owner, operator string) error {
	msg := bson.M{"owner": owner, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventQuote(uid, quote, operator string) error {
	msg := bson.M{"quote": quote, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventTargets(uid, operator string, targets []string) error {
	msg := bson.M{"targets": targets, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventInfo(uid, name, desc, operator string) error {
	msg := bson.M{"name": name, "desc": desc, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventTags(uid, operator string, list []string) error {
	msg := bson.M{"tags": list, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventAssets(uid, operator string, list []string) error {
	msg := bson.M{"assets": list, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableEvent, uid, msg)
	return err
}

func UpdateEventCover(uid, operator, cover string) error {
	msg := bson.M{"cover": cover, "operator": operator, TimeUpdated: time.Now().Unix()}
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
