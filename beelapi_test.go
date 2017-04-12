package beelapi

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

var client APIClient
var records []CallRecord
var rec CallRecord

func init() {

	client = NewApiClient("token")
}

// TestGetRecords Тест на получение информации о записях
func TestGetRecords(t *testing.T) {
	httpmock.Activate()
	defer httpmock.Deactivate()
	resp := CallRecord{}
	resp.Id = "1"
	resp.Abonent.Phone = "0000000000"
	resp.Date = time.Until(time.Now())
	resp.Direction = "OUTBOUND"
	resp.Duration = 100000
	resp.FileSize = 200000
	RegisterJsonDataMock("GET", client.BaseApiUrl+"records", resp)
	records, err := client.GetRecords(0)
	fireError(err, "Не удалось получить инфо о записях. ")
	rec = records[0]
	if rec.Abonent != resp.Abonent {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверен номер абонента. Ожидалось %s получено %s", resp.Abonent, rec.Abonent)
	}
	if rec.Date != resp.Date {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверна дата звонка. Ожидалось %s получено %s", resp.Date, rec.Date)
	}
	if rec.Direction != resp.Direction {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверено направление звонка. Ожидалось %s получено %s", resp.Direction, rec.Direction)
	}
	if rec.Duration != resp.Duration {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверена продолжительность звонка. Ожидалось %s получено %s", resp.Duration, rec.Duration)
	}
	if rec.FileSize != resp.FileSize {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверен размер файла. Ожидалось %s получено %s", resp.FileSize, rec.FileSize)
	}
	if rec.Id != resp.Id {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверен ID записи. Ожидалось %s получено %s", resp.Id, rec.Id)
	}
	if rec.Phone != resp.Phone {
		log.Fatalf("Ошибка при проверке ответа на запрос о получении инфо о записях. Неверен номер клиента. Ожидалось %s получено %s", resp.Phone, rec.Phone)
	}
}

//  TestGetWavFileFromServer Тест на получение файла с сервера
func TestGetWavFileFromServer(t *testing.T) {
	httpmock.Activate()
	defer httpmock.Deactivate()
	recId := rec.Id
	url := fmt.Sprintf("%s/v2/records/%s/download", client.BaseApiUrl, recId)
	file, err := os.Open("test.wav")
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

//  TestDeleteRecord Тест на конвертацию из wav файла в mp3
func TestDeleteRecord(t *testing.T) {
	err := client.DeleteRecord(rec.Id)
	fireError(err, "")
}

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
func RegisterChunkedDataMock(method string, url string, r []byte) {
	httpmock.RegisterResponder(method, url,
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewBytesResponse(200, r)
			return resp, nil
		})
}
