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
	Key   string     `json:"key" bson:"key"` // 属性UID
	Words []WordInfo `json:"values" bson:"values"`
}

type WordInfo struct {
	// 实体UID,如果存在说明该属性对应一个实体
	UID string `json:"uid" bson:"uid"`
	// 属性值
	Name string `json:"name" bson:"name"`
}

type PairInfo struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

type RelationCaseInfo struct {
	UID       string `json:"uid" bson:"uid"`
	Direction uint8  `json:"direction" bson:"direction"`
	Name      string `json:"relation" bson:"relation"` // 如果是定制则是显示定制名称，否则显示类型名称
	Category  string `json:"category" bson:"category"` //关系类型UID
	Entity    string `json:"entity" bson:"entity"`     //对应实体UID
	Weight    uint32 `json:"weight" bson:"weight"` //亲密度或者权重
}

type EventBrief struct {
	Name        string
	Description string // 描述
	Quote       string // 引用或者备注
	Date        DateInfo
	Place       PlaceInfo
	Tags        []string
	Assets      []string
}

type Date struct {
	Type  uint8  `json:"type" bson:"type"` // AD or BC
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

func (mine *PropertyInfo) HadWordByValue(val string) bool {
	for i := 0; i < len(mine.Words); i += 1 {
		if mine.Words[i].Name == val {
			return true
		}
	}
	return false
}

func (mine *Date) String() string {
	if mine.Type > 0 {
		return fmt.Sprintf("%d/%d/%d", mine.Year, mine.Month, mine.Day)
	} else {
		return fmt.Sprintf("-%d/%d/%d", mine.Year, mine.Month, mine.Day)
	}
}

func (mine *Date) Parse(msg string) error {
	if len(msg) < 1 {
		return errors.New("the date is empty")
	}
	if strings.Contains(msg, "-") {
		mine.Type = 0
	} else {
		mine.Type = 1
	}
	mine.Name = msg
	array := strings.Split(msg, "/")
	if array != nil && len(array) > 2 {
		year, _ := strconv.ParseUint(array[0], 10, 32)
		mine.Year = uint16(year)
		month, _ := strconv.ParseUint(array[1], 10, 32)
		mine.Month = uint8(month)
		day, _ := strconv.ParseUint(array[2], 10, 32)
		mine.Day = uint8(day)
		return nil
	} else {
		return errors.New("the split array is nil")
	}
}
