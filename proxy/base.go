package proxy

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

/**
实体填写的属性值
*/
type PropertyInfo struct {
	Key   string     `json:"key" bson:"key"`
	Words []WordInfo `json:"values" bson:"values"`
}

type WordInfo struct {
	//实体UID,如果存在说明该属性对应一个实体
	UID  string `json:"key" bson:"key"`
	Name string `json:"value" bson:"value"`
}

type PairInfo struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

type RelationInfo struct {
	UID 	   string `json:"uid" bson:"uid"`
	Direction  uint8  `json:"direction" bson:"direction"`
	Name       string `json:"relation" bson:"relation"`
	Category   string `json:"category" bson:"category"`
	Entity  	string `json:"entity" bson:"entity"`
}

type Date struct {
	Type uint8 `json:"type" bson:"type"` // AD or BC
	Day   uint8  `json:"day" bson:"day"`
	Month uint8  `json:"month" bson:"month"`
	Year  uint16 `json:"year" bson:"year"`
	Name  string `json:"name" bson:"name"`
}

type DateInfo struct {
	UID   string `json:"uid" bson:"uid"` //实体UID
	Name  string `json:"name" bson:"name"`
	Begin Date   `json:"begin" bson:"begin"`
	End   Date   `json:"end" bson:"end"`
}

type PlaceInfo struct {
	UID      string `json:"uid" bson:"uid"` //实体UID
	Name     string `json:"name" bson:"name"`
	Location string `json:"location" bson:"location"`
}

type EventPoint struct {
	ID          uint64         `json:"id" bson:"id"`
	Description string         `json:"desc" bson:"desc"`
	Date        DateInfo       `json:"date" bson:"date"`
	Place       PlaceInfo      `json:"place" bson:"place"`
	Assets      []string       `json:"assets" bson:"assets"`
	Relations   []RelationInfo `json:"relations" bson:"relations"`
}

type Location struct {
	Name      string  `json:"name" bson:"name"`
	Longitude float32 `json:"longitude" bson:"longitude"`
	Latitude  float32 `json:"latitude" bson:"latitude"`
}

func (mine *PropertyInfo) HadWordByEntity(uid string) bool {
	for i := 0; i < len(mine.Words); i += 1 {
		if mine.Words[i].UID == uid {
			return true
		}
	}
	return false
}

func (mine *Date)String() string {
	if mine.Type > 0 {
		return fmt.Sprintf("%d/%d/%d", mine.Year, mine.Month, mine.Day)
	}else{
		return fmt.Sprintf("-%d/%d/%d", mine.Year, mine.Month, mine.Day)
	}
}

func (mine *Date)Parse(msg string) error {
	if len(msg) < 1 {
		return errors.New("the date is empty")
	}
	if strings.Contains(msg, "-") {
		mine.Type = 0
	}else{
		mine.Type = 1
	}
	array := strings.Split(msg, "/")
	if array != nil && len(array) > 2 {
		year,_ := strconv.ParseUint(array[0], 10, 32)
		mine.Year = uint16(year)
		month,_ := strconv.ParseUint(array[1], 10, 32)
		mine.Month = uint8(month)
		day,_ := strconv.ParseUint(array[2], 10, 32)
		mine.Day = uint8(day)
		return nil
	}else{
		return errors.New("the split array is nil")
	}
}
