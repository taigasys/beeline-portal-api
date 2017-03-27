package beelapi

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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

// GetRecord метод возвращает ифно об записей по индексу
func (r *RecordInfos) GetRecord(i int64) *RecordInfo {
	return &r.CallInfos[i]
}

// RecordInfos структура для хранения краткой информации о записях, необхадима для отправки запросо на получение файлов
type ShortRecordInfos struct {
	ShortCallInfos []ShortRecordInfo
	Count          int
}

// RecordInfo структура для хранения краткой информации об отдельной записи
type ShortRecordInfo struct {
	RecordId int64
	Status   string
	SaveDate time.Time
}

// IRecordsInfoProvider интерфейс для отправки запросов на получение файлов разговоров
type IRecordsInfoProvider interface {
	Len() int64
	GetRecordInfo(index int64) IRecordInfoProvider
}

// IRecordInfoProvider интерфейс для отправки запросов на получение файлов разговоров, хранит инфо об отдельной записи
type IRecordInfoProvider interface {
	GetId() int64
	GetStatus() string
	SetStatus(s string)
}

// Len метод расситывает количество записей
func (r *RecordInfos) Len() int64 {
	return int64(len(r.CallInfos))
}

// GetRecordInfo метод возвращает инфо об отдельной записи, в данном случае - полную
func (r *RecordInfos) GetRecordInfo(index int64) IRecordInfoProvider {
	return &r.CallInfos[index]
}

// GetId метод возвращает ID записи разговора
func (r *RecordInfo) GetId() int64 {
	return r.RecordId
}

// GetStatus метод возвращает статус хранения записи разговора
func (r *RecordInfo) GetStatus() string {
	return r.Status
}

// SetStatus метод устанавливает статус хранения записи разговора
func (r *RecordInfo) SetStatus(s string) {
	r.Status = s
}

// GetId метод возвращает ID записи разговора
func (ri *ShortRecordInfo) GetId() int64 {
	return ri.RecordId
}

// GetStatus метод возвращает статус хранения записи разговора
func (ri *ShortRecordInfo) GetStatus() string {
	return ri.Status
}

// SetStatus метод устанавливает статус хранения записи разговора
func (ri *ShortRecordInfo) SetStatus(s string) {
	ri.Status = s
}

// Len  метод расситывает количество записей
func (ri *ShortRecordInfos) Len() int64 {
	return int64(len(ri.ShortCallInfos))
}

// GetRecordInfo метод возвращает инфо об отдельной записи, в данном случае - краткую
func (ri *ShortRecordInfos) GetRecordInfo(index int64) IRecordInfoProvider {
	return &ri.ShortCallInfos[index]
}

//BAPIError Тип хранения ошибок
type BAPIError struct {
	Msg string
}

func (d BAPIError) Error() string {
	return d.Msg
}

// Текущая дата
var EndDate time.Time

// Дата, с которой получаем записи
var StartDate time.Time

// BuildXMLRequest Подготавливает тело запроса на получение информации о записях на сервере Билайн
// dir - аргумент типа звонков(входящие или исхордящие)
func BuildXMLRequest(dir string) string {
	PageSize = 200
	EndDate = time.Now()
	StartDate = time.Now().Add(-time.Duration(PeriodInHours) * 60 * time.Minute)
	return `<?xml version="1.0" encoding="utf-8"?>
	<tns:ListCallRecordRequest xmlns:tns="http://client.pub.api.cloudpbx.beeline.ru">
	<pageNumber>0</pageNumber>
	<pageSize>` + strconv.Itoa(PageSize) + `</pageSize>
	<direction>` + dir + `</direction>
	<dateFrom>` + StartDate.Format(time.RFC3339) + `</dateFrom>
	<dateTo>` + EndDate.Format(time.RFC3339) + `</dateTo>
	<sort>
	<direction>ASC</direction>
	<field>Date</field>
	</sort>
	</tns:ListCallRecordRequest>`
}

// GetRecordsInfoFromServer Получает информацию о количестве записей на сервере Билайн за данных период
func GetRecordsInfoFromServer(r string) (*RecordInfos, error) {

	reqBody := bytes.NewBufferString(r)
	req, err := http.NewRequest("PUT", RecordListUrl, reqBody)
	if err != nil {
		return nil, BAPIError{Msg: "Ошибка при подготовке запроса к серверу Beeline на получение информации о файлах:" + err.Error()}
	}

	req.Header.Set("Content-Type", "application/xml")
	req.SetBasicAuth(Username, Passwrd)
	cl := &http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, BAPIError{Msg: "Возникла ошибка при отправке запроса: " + err.Error()}
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, BAPIError{Msg: "Возникла ошибка! Получен ответ от сервера: " + resp.Status + " Код ответа: " + strconv.Itoa(resp.StatusCode)}
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, BAPIError{Msg: "Ошибка при обработке ответа от сервера: " + err.Error()}
	}
	body := bytes.NewBufferString(string(bodyBytes))
	var rs RecordInfos
	xmlBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, BAPIError{Msg: "Ошибка при обработке ответа от сервера: " + err.Error()}
	}
	err = xml.Unmarshal(xmlBytes, &rs)
	if err != nil {
		return nil, BAPIError{Msg: "Ошибка при обработке ответа от сервера: " + err.Error()}
	}
	if rs.Count == 0 {
		log.Println("На сервере не найдено записей разговоров за указанный период. Программа завершила работу")
		return nil, nil
	} else {
		log.Printf("На сервере найдено %d записи(ей) разговоров за указанный период...", rs.Count)
	}
	return &rs, nil
}
