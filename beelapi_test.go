package beelineapi

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	httpmock "gopkg.in/jarcoal/httpmock.v1"

	"time"
)

//Настройки клиента
var client APIClient

//Список записей
var records []CallRecord

//Отдельная тестируемая запись
var rec CallRecord

func init() {

	client = NewApiClient("token")
}

// TestGetRecords Тест на получение информации о записях
func TestGetRecords(t *testing.T) {
	httpmock.Activate()
	defer httpmock.Deactivate()
	testRecs := []CallRecord{}
	testRec := CallRecord{}
	testRec.Id = "test"
	testRec.Abonent.Phone = "0000000000"
	testRec.Date = time.Until(time.Now())
	testRec.Direction = "OUTBOUND"
	testRec.Duration = 100000
	testRec.FileSize = 200000
	testRecs = append(testRecs, testRec)
	RegisterJsonDataMock("GET", client.BaseApiUrl+"records", testRecs)
	records, err := client.GetRecords(0)
	fireError(err, "Не удалось получить инфо о записях. ")
	rec = records[0]
	// Сравниваем результаты ответа и заполненной структуры
	if rec.Abonent != testRec.Abonent {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверен номер абонента. Ожидалось %s получено %s", testRec.Abonent, rec.Abonent)
	}
	if rec.Date != testRec.Date {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверна дата звонка. Ожидалось %s получено %s", testRec.Date, rec.Date)
	}
	if rec.Direction != testRec.Direction {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверено направление звонка. Ожидалось %s получено %s", testRec.Direction, rec.Direction)
	}
	if rec.Duration != testRec.Duration {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверена продолжительность звонка. Ожидалось %s получено %s", testRec.Duration, rec.Duration)
	}
	if rec.FileSize != testRec.FileSize {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверен размер файла. Ожидалось %s получено %s", testRec.FileSize, rec.FileSize)
	}
	if rec.Id != testRec.Id {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверен ID записи. Ожидалось %s получено %s", testRec.Id, rec.Id)
	}
	if rec.Phone != testRec.Phone {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверен номер клиента. Ожидалось %s получено %s", testRec.Phone, rec.Phone)
	}
}

//  TestGetWavFileFromServer Тест на получение файла с сервера
func TestGetWavFileFromServer(t *testing.T) {
	httpmock.Activate()
	defer httpmock.Deactivate()
	recId := rec.Id
	url := fmt.Sprintf("%sv2/records/%s/download", client.BaseApiUrl, recId)
	fmt.Println(url)
	file, err := os.Open("test/test.wav")
	fireError(err, "Тестовый файл с записью не удалось открыть")
	resp, err := ioutil.ReadAll(file)
	fireError(err, "Тестовый файл с записью не удалось считать")
	RegisterChunkedDataMock("GET", url, resp)

	reader, err := client.GetRecordFile(recId)
	fireError(err, "")
	contentOfFile, err := ioutil.ReadAll(reader)
	fireError(err, "Не удалось считать данные из потока с файлом записи. ")
	if len(contentOfFile) == 0 {
		log.Fatalln("Из потока с файлом записи от сервера получен пустой массив данных")
	}
}

//  TestDeleteRecord Тест на удаление записи с сервера Билайн
func TestDeleteRecord(t *testing.T) {
	httpmock.Activate()
	defer httpmock.Deactivate()
	url := fmt.Sprintf("%sv2/records/%s", client.BaseApiUrl, rec.Id)
	RegisterJsonDataMock("DELETE", url, nil)
	err := client.DeleteRecord(rec.Id)
	fireError(err, "")
}

//   RegisterJsonDataMock Добавление обработчика к имитатору сервера url на запрос инфо о записях
func RegisterJsonDataMock(method string, url string, r interface{}) {
	httpmock.RegisterResponder(method, url,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, r)
			if err != nil {
				return httpmock.NewStringResponse(500, err.Error()), nil
			}
			return resp, nil
		})
}

//  RegisterChunkedDataMock Добавление обработчика к имитатору сервера url на запрос получении файла записи
func RegisterChunkedDataMock(method string, url string, r []byte) {
	httpmock.RegisterResponder(method, url,
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewBytesResponse(200, r)
			return resp, nil
		})
}
