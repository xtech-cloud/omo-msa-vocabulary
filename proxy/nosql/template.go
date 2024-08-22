package nosql

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Template struct {
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

	Name        string `json:"name" bson:"name"`
	Type        uint8 `json:"type" bson:"type"`
	SkipRows    int32 `json:"skipRows" bson:"skipRows"`
    Columns     map[string]int32;
	Url         string   `json:"url" bson:"url"`
	Comments    string  `json:"comments" bson:"comments"`
}


func GetTemplateNextID() uint64 {
	num, _ := getSequenceNext(TableTemplate)
	return num
}

func CreateTemplate(info *Template) error {
	_, err := insertOne(TableTemplate, info)
	if err != nil {
		return err
	}
	return nil
}

func GetTemplates() ([]*Template, error) {
	var items = make([]*Template, 0, 20)
	filter := bson.M{TimeDeleted: 0}
	cursor, err1 := findMany(TableTemplate, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Template)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetTemplate(uid string) (*Template, error) {
	result, err := findOne(TableTemplate, uid)
	if err != nil {
		return nil, err
	}
	model := new(Template)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}


func GetTemplateByName(name string) (*Template, error) {
	msg := bson.M{"name": name, TimeDeleted: 0}
	result, err := findOneBy(TableTemplate, msg)
	if err != nil {
		return nil, err
	}
	model := new(Template)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func RemoveTemplate(uid, operator string) error {
	_, err := removeOne(TableTemplate, uid, operator)
	return err
}

func UpdateTemplateBase(uid string, name string, _type uint8, skipRows int32, columns map[string]int32, url, comments, operator string) error {
	msg := bson.M{"name": name, "type": _type, "skipRows": skipRows, "columns": columns, "url": url, 
		"comments": comments, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableTemplate, uid, msg)
	return err
}

