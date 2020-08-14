package cache

import (
	"omo.msa.vocabulary/proxy"
)

const (
	DirectionTypeDouble DirectionType = 0
	DirectionTypeFromTo DirectionType = 1
	DirectionTypeToFrom DirectionType = 2
)

const (
	LinkTypeEmpty LinkType = "Other"
	LinkTypePersons  LinkType = "Persons"
	LinkTypeEvents  LinkType = "Events"
	LinkTypeInhuman	  LinkType = "Inhuman"      //
)

type LinkType string
type DirectionType uint8

type LinkInfo struct {
	Direction DirectionType
	ID int64
	Name string
	Label string
	Relation string
	From string
	To   string
}

func RemoveLink(id int64) error {
	var err error
	var link *proxy.Link
	link,err = proxy.GetLinkByID(id)
	if err == nil {
		err = link.Delete()
	}
	return err
}

func (mine *LinkInfo)initInfo(db *proxy.Link, from, to string) bool {
	if db == nil {
		return false
	}
	mine.ID = db.ID
	mine.Name = db.Name
	mine.From = from
	mine.To = to
	mine.Label = db.Label
	mine.Relation = db.Relation
	mine.Direction = DirectionType(db.Direction)
	return true
}

func (mine *LinkInfo)HadNode(uid string) bool {
	if mine.From == uid {
		return true
	}else if mine.To == uid {
		return true
	}else {
		return false
	}
}

func (mine *LinkInfo)HadAll(from, to string) bool {
	if (mine.From == from || mine.To == from) &&
		(mine.From == to || mine.To == to) {
		return true
	}
	return false
}

func (mine *LinkInfo)GetAnotherNode(uid string) string {
	if mine.From == uid {
		return mine.To
	}else if mine.To == uid {
		return mine.From
	}else {
		return ""
	}
}