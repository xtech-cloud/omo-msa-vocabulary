package cache

import (
	"omo.msa.vocabulary/proxy/nosql"
)

type AddressInfo struct {
	UID		string `json:"-"`
	ID		uint64	`json:"-"`
	Country string	`json:"-"`
	Province string	`json:"-"`
	City	 string	`json:"-"`
	District string	`json:"-"`
	Town	string	`json:"-"`
	Village string	`json:"-"`
	Street	string	`json:"-"`
	Number	string `json:"-"`
}

func (mine *AddressInfo) initInfo(info *nosql.Address) bool {
	if info == nil {
		return false
	}
	mine.UID = info.UID.Hex()
	mine.ID = info.ID
	mine.Country = info.Country
	mine.Province = info.Province
	mine.City = info.City
	mine.District = info.District
	mine.Town = info.Town
	mine.Village = info.Village
	mine.Street =  info.Street
	mine.Number = info.Number

	return true
}

func (mine *AddressInfo)String() string {
	return mine.Province + " " + mine.City
}
