package beelapi

import (
	"encoding/xml"
	"time"
)

var (
	Username      string
	Passwrd       string
	PeriodInHours int
	FtpLogin      string
	FtpPsw        string
	RecordListUrl string
	Provider      string
	DBTable       string
)

// Количество файлов, информация о которых возвращается нашей утилите
var PageSize int

// Records структура хранения подробной информации о записях и их количестве
type RecordInfos struct {
	XMLName   xml.Name     `xml:"ListCallRecordResponse"`
	CallInfos []RecordInfo `xml:"CallRecord"`
	Count     int          `xml:"totalRecordQuantity"`
}

// Record структура хранения подробной информации об отдельной записи
type RecordInfo struct {
	Id            int       `xml:"-"`
	RecordId      int64     `xml:"recordId"`
	InternalId    int64     `xml:"-"`
	AbonentPhone  string    `xml:"Abonent>phone"`
	ClientPhone   string    `xml:"phone"`
	CallDirection string    `xml:"callDirection"`
	CallDate      time.Time `xml:"date"`
	Duration      int64     `xml:"duration"`
	FileSize      int64     `xml:"-"`
	Status        string    `xml:"-"`
	SaveDate      time.Time `xml:"-"`
	Provider      string    `xml:"-"`
}
