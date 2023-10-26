package nosql

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const timeOut = 10 * time.Second

const (
	TimeCreated = "created"
	TimeUpdated = "updated"
	TimeDeleted = "deleted"
)

func UpdateItemTime(table, uid string, created, updated, del time.Time) {
	d := del.Unix()
	if d < 0 {
		d = 0
	}
	u := updated.Unix()
	if u < 0 {
		u = 0
	}
	msg := bson.M{TimeCreated: created.Unix(), TimeUpdated: u, TimeDeleted: d}
	_, _ = updateOne(table, uid, msg)
}

func GetAll[T any](table string, items []*T) []*T {
	cursor, err1 := findAll(table, 0)
	if err1 != nil {
		return make([]*T, 0, 1)
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var node = new(T)
		if err := cursor.Decode(node); err != nil {
			return make([]*T, 0, 1)
		} else {
			items = append(items, node)
		}
	}
	return items
}

func findAll(collection string, limit int64) (*mongo.Cursor, error) {
	if len(collection) < 1 {
		return nil, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return nil, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	filter := bson.M{}
	var cursor *mongo.Cursor
	var err error
	if limit > 0 {
		cursor, err = c.Find(ctx, filter, options.Find().SetSort(bson.M{TimeCreated: -1}).SetLimit(limit))
	} else {
		cursor, err = c.Find(ctx, filter)
	}
	if err != nil {
		return nil, err
	}
	return cursor, nil
}

func insertOne(collection string, info interface{}) (interface{}, error) {
	if len(collection) < 1 {
		return "", errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return "", errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	result, err := c.InsertOne(ctx, info)
	if err != nil {
		return "", err
	}
	return result.InsertedID, nil
}

func getCount(collection string) (int64, error) {
	if len(collection) < 1 {
		return 0, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return 0, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	result, err := c.EstimatedDocumentCount(ctx, nil)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func getTotalCount(collection string) (int64, error) {
	if len(collection) < 1 {
		return 0, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return 0, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	result, err := c.EstimatedDocumentCount(ctx, nil)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func getCountBy(collection string, filter bson.M) (int64, error) {
	if len(collection) < 1 {
		return 0, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return 0, errors.New("can not found the collection of" + collection)
	}
	opts := options.Count().SetMaxTime(time.Second * 2)
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	result, err := c.CountDocuments(ctx, filter, opts)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func deleteOne(collection, uid string) (int64, error) {
	if len(collection) < 1 {
		return 0, errors.New("the collection is empty")
	}
	if len(uid) < 2 {
		return 0, errors.New("the uid is empty of " + collection)
	}
	objID, e := primitive.ObjectIDFromHex(uid)
	if e != nil {
		return 0, e
	}
	c := noSql.Collection(collection)
	if c == nil {
		return 0, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	node := bson.M{"_id": objID}
	result, err := c.DeleteOne(ctx, node)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func removeOne(collection, uid, operator string) (int64, error) {
	if len(collection) < 1 {
		return 0, errors.New("the collection is empty")
	}
	if len(uid) < 2 {
		return 0, errors.New("the uid is empty of " + collection)
	}
	objID, e := primitive.ObjectIDFromHex(uid)
	if e != nil {
		return 0, e
	}
	c := noSql.Collection(collection)
	if c == nil {
		return 0, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	filter := bson.M{"_id": objID}
	node := bson.M{"$set": bson.M{"operator": operator, TimeDeleted: time.Now().Unix()}}
	result, err := c.UpdateOne(ctx, filter, node)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func hadOne(collection string, filter bson.M) (bool, error) {
	if len(collection) < 1 {
		return false, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return false, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	result := c.FindOne(ctx, filter)
	if result.Err() != nil {
		return false, result.Err()
	}
	return true, nil
}

func updateOne(collection, uid string, data bson.M) (int64, error) {
	if len(collection) < 1 {
		return 0, errors.New("the collection is empty")
	}
	if len(uid) < 2 {
		return 0, errors.New("the uid is empty of " + collection)
	}
	objID, e := primitive.ObjectIDFromHex(uid)
	if e != nil {
		return 0, e
	}
	c := noSql.Collection(collection)
	if c == nil {
		return 0, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	filter := bson.M{"_id": objID}
	node := bson.M{"$set": data}
	result, err := c.UpdateOne(ctx, filter, node)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

/**
往数组里面追加一个元素
*/
func appendElement(collection, uid string, data bson.M) (int64, error) {
	if len(collection) < 1 {
		return 0, errors.New("the collection is empty")
	}
	if len(uid) < 2 {
		return 0, errors.New("the uid is empty of " + collection)
	}
	objID, e := primitive.ObjectIDFromHex(uid)
	if e != nil {
		return 0, e
	}
	c := noSql.Collection(collection)
	if c == nil {
		return 0, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	filter := bson.M{"_id": objID}
	node := bson.M{"$push": data, "$set": bson.M{TimeUpdated: time.Now().Unix()}}
	result, err := c.UpdateOne(ctx, filter, node)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

/**
从数组里面移除一个元素
*/
func removeElement(collection, uid string, data bson.M) (int64, error) {
	if len(collection) < 1 {
		return 0, errors.New("the collection is empty")
	}
	if len(uid) < 2 {
		return 0, errors.New("the uid is empty of " + collection)
	}
	objID, e := primitive.ObjectIDFromHex(uid)
	if e != nil {
		return 0, e
	}
	c := noSql.Collection(collection)
	if c == nil {
		return 0, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	filter := bson.M{"_id": objID}
	node := bson.M{"$pull": data, "$set": bson.M{TimeUpdated: time.Now().Unix()}}
	result, err := c.UpdateOne(ctx, filter, node)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func updateOneBy(collection string, filter bson.M, update bson.M) (int64, error) {
	if len(collection) < 1 {
		return 0, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return 0, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	result, err := c.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func findOne(collection, uid string) (*mongo.SingleResult, error) {
	if len(collection) < 2 {
		return nil, errors.New("the collection is empty")
	}
	if len(uid) < 2 {
		return nil, errors.New("the uid is empty of " + collection)
	}
	objID, e := primitive.ObjectIDFromHex(uid)
	if e != nil {
		return nil, e
	}
	c := noSql.Collection(collection)
	if c == nil {
		return nil, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	filter := bson.M{"_id": objID}
	result := c.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	return result, nil
}

func findOneBy(collection string, filter bson.M) (*mongo.SingleResult, error) {
	if len(collection) < 2 {
		return nil, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return nil, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	result := c.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	return result, nil
}

func findOneOfField(collection, uid string, selector bson.M) (*mongo.SingleResult, error) {
	if len(collection) < 2 {
		return nil, errors.New("the collection is empty")
	}
	if len(uid) < 2 {
		return nil, errors.New("the uid is empty of " + collection)
	}
	objID, e := primitive.ObjectIDFromHex(uid)
	if e != nil {
		return nil, e
	}
	c := noSql.Collection(collection)
	if c == nil {
		return nil, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	filter := bson.M{"_id": objID}
	result := c.FindOne(ctx, filter, options.FindOne().SetProjection(selector))
	if result.Err() != nil {
		return nil, result.Err()
	}
	return result, nil
}

func findOneByOpt(collection string, filter bson.M, selector bson.M) (*mongo.SingleResult, error) {
	if len(collection) < 2 {
		return nil, errors.New("the uid is empty of " + collection)
	}
	c := noSql.Collection(collection)
	if c == nil {
		return nil, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	result := c.FindOne(ctx, filter, options.FindOne().SetProjection(selector))
	if result.Err() != nil {
		return nil, result.Err()
	}
	return result, nil
}

func findMany(collection string, filter bson.M, limit int64) (*mongo.Cursor, error) {
	if len(collection) < 1 {
		return nil, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return nil, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	var cursor *mongo.Cursor
	var err error
	if limit > 0 {
		cursor, err = c.Find(ctx, filter, options.Find().SetSort(bson.M{TimeCreated: -1}).SetLimit(limit))
	} else {
		cursor, err = c.Find(ctx, filter)
	}
	if err != nil {
		return nil, err
	}
	return cursor, nil
}

func getCountByFilter(collection string, filter bson.M) (int64, error) {
	if len(collection) < 1 {
		return -1, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return -1, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	return c.CountDocuments(ctx, filter)
}

func findManyByOpts(collection string, filter bson.M, opts *options.FindOptions) (*mongo.Cursor, error) {
	if len(collection) < 1 {
		return nil, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return nil, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	cursor, err := c.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	return cursor, nil
}

func findAllByOpts(collection string, opts *options.FindOptions) (*mongo.Cursor, error) {
	if len(collection) < 1 {
		return nil, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return nil, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	filter := bson.M{TimeDeleted: 0}
	cursor, err := c.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	return cursor, nil
}

func findAllEnable(collection string, limit int64) (*mongo.Cursor, error) {
	if len(collection) < 1 {
		return nil, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return nil, errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	filter := bson.M{TimeDeleted: 0}
	var cursor *mongo.Cursor
	var err error
	if limit > 0 {
		cursor, err = c.Find(ctx, filter, options.Find().SetSort(bson.M{TimeCreated: -1}).SetLimit(limit))
	} else {
		cursor, err = c.Find(ctx, filter)
	}
	if err != nil {
		return nil, err
	}
	return cursor, nil
}

func dropOne(collection string) error {
	if len(collection) < 1 {
		return errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return errors.New("can not found the collection of" + collection)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	err := c.Drop(ctx)
	if err != nil {
		return err
	}

	return nil
}

func copyOne(collection string) (*mongo.Collection, error) {
	if len(collection) < 1 {
		return nil, errors.New("the collection is empty")
	}
	c := noSql.Collection(collection)
	if c == nil {
		return nil, errors.New("can not found the collection of" + collection)
	}
	tmp, err := c.Clone()
	if err != nil {
		return nil, err
	}
	return tmp, nil
}
