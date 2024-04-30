package nosql

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"omo.msa.vocabulary/proxy"
	"time"
)

type Box struct {
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

	Name      string               `json:"name" bson:"name"`
	Type      uint8                `json:"type" bson:"type"`
	Cover     string               `json:"cover" bson:"cover"`
	Remark    string               `json:"remark" bson:"remark"`
	Owner     string               `json:"owner" bson:"owner"`
	Concept   string               `json:"concept" bson:"concept"`
	Workflow  string               `json:"workflow" bson:"workflow"`
	Keywords  []string             `json:"keywords" bson:"keywords"`
	Users     []string             `json:"users" bson:"users"`
	Reviewers []string             `json:"reviewers" bson:"reviewers"`
	Contents  []*proxy.ContentInfo `json:"contents" bson:"contents"`
}

func CreateBox(info *Box) error {
	_, err := insertOne(TableBox, info)
	if err != nil {
		return err
	}
	return nil
}

func GetBoxNextID() uint64 {
	num, _ := getSequenceNext(TableBox)
	return num
}

func GetBox(uid string) (*Box, error) {
	result, err := findOne(TableBox, uid)
	if err != nil {
		return nil, err
	}
	model := new(Box)
	err1 := result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetBoxes() ([]*Box, error) {
	var items = make([]*Box, 0, 20)
	cursor, err1 := findAllEnable(TableBox, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Box)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func HadBoxByName(name string) (bool, error) {
	msg := bson.M{"name": name}
	return hadOne(TableBox, msg)
}

func GetBoxesByType(owner string, tp uint8) ([]*Box, error) {
	var items = make([]*Box, 0, 20)
	filter := bson.M{"owner": owner, "type": tp, TimeDeleted: 0}
	cursor, err1 := findMany(TableBox, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Box)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetBoxesByOwner(owner string) ([]*Box, error) {
	var items = make([]*Box, 0, 20)
	filter := bson.M{"owner": owner, TimeDeleted: 0}
	cursor, err1 := findMany(TableBox, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Box)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetBoxCount() int64 {
	filter := bson.M{TimeDeleted: 0}
	num, err1 := getCount2(TableBox, filter)
	if err1 != nil {
		return num
	}

	return num
}

func GetBoxByName(name string) (*Box, error) {
	filter := bson.M{"name": name, TimeDeleted: 0}
	result, err1 := findOneBy(TableBox, filter)
	model := new(Box)
	err1 = result.Decode(model)
	if err1 != nil {
		return nil, err1
	}
	return model, nil
}

func GetBoxesByUser(user string) ([]*Box, error) {
	var items = make([]*Box, 0, 20)
	filter := bson.M{"users": user, TimeDeleted: 0}
	cursor, err1 := findMany(TableBox, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Box)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetBoxesByReviewer(user string) ([]*Box, error) {
	var items = make([]*Box, 0, 20)
	filter := bson.M{"reviewers": user, TimeDeleted: 0}
	cursor, err1 := findMany(TableBox, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Box)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetBoxesByRegex(key, val string) ([]*Box, error) {
	msg := bson.M{key: bson.M{"$regex": val}, TimeDeleted: 0}
	cursor, err1 := findMany(TableBox, msg, 0)
	if err1 != nil {
		return nil, err1
	}
	var items = make([]*Box, 0, 100)
	for cursor.Next(context.Background()) {
		var node = new(Box)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetBoxesByPage(start, num int64) ([]*Box, error) {
	filter := bson.M{TimeDeleted: 0}
	opts := options.Find().SetSort(bson.D{{TimeCreated, -1}}).SetLimit(num).SetSkip(start)
	cursor, err1 := findManyByOpts(TableBox, filter, opts)
	if err1 != nil {
		return nil, err1
	}
	var items = make([]*Box, 0, 20)
	for cursor.Next(context.TODO()) {
		var node = new(Box)
		if err := cursor.Decode(&node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetBoxesByKeyword(key string) ([]*Box, error) {
	var items = make([]*Box, 0, 20)
	filter := bson.M{"contents.keyword": key, TimeDeleted: 0}
	cursor, err1 := findMany(TableBox, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Box)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func GetBoxesByConcept(concept string) ([]*Box, error) {
	var items = make([]*Box, 0, 20)
	filter := bson.M{"concept": concept, TimeDeleted: 0}
	cursor, err1 := findMany(TableBox, filter, 0)
	if err1 != nil {
		return nil, err1
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(Box)
		if err := cursor.Decode(node); err != nil {
			return nil, err
		} else {
			items = append(items, node)
		}
	}
	return items, nil
}

func UpdateBoxBase(uid, name, desc, operator string) error {
	msg := bson.M{"name": name, "remark": desc, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableBox, uid, msg)
	return err
}

func UpdateBoxCover(uid string, icon string) error {
	msg := bson.M{"cover": icon, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableBox, uid, msg)
	return err
}

func UpdateBoxConcept(uid, operator, concept string) error {
	msg := bson.M{"operator": operator, "concept": concept, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableBox, uid, msg)
	return err
}

func UpdateBoxOwner(uid, owner string) error {
	msg := bson.M{"owner": owner, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableBox, uid, msg)
	return err
}

func UpdateBoxContents(uid, operator string, list []*proxy.ContentInfo) error {
	msg := bson.M{"contents": list, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableBox, uid, msg)
	return err
}

func RemoveBox(uid, operator string) error {
	_, err := removeOne(TableBox, uid, operator)
	return err
}

func UpdateBoxUsers(uid, operator string, list []string) error {
	msg := bson.M{"users": list, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableBox, uid, msg)
	return err
}

func UpdateBoxReviewers(uid, operator string, list []string) error {
	msg := bson.M{"reviewers": list, "operator": operator, TimeUpdated: time.Now().Unix()}
	_, err := updateOne(TableBox, uid, msg)
	return err
}

func AppendBoxKeyword(uid string, key string) error {
	msg := bson.M{"keywords": key}
	_, err := appendElement(TableBox, uid, msg)
	return err
}

func SubtractBoxKeyword(uid string, key string) error {
	msg := bson.M{"keywords": key}
	_, err := removeElement(TableBox, uid, msg)
	return err
}
