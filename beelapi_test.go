package beelapi

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

func init() {
	PeriodInHours = 4
	RecordListUrl = "https://cloudpbx.beeline.ru/api/pub/client/call/record/list"
	RecordFileUrl = "https://cloudpbx.beeline.ru/api/pub/client/call/record/list"
	DBTable = "call_record"
}

// TestGetClientsFromDB Тест на выборку данных
func TestGetRecordsInfoFromServer(t *testing.T) {
	num := 1
	bodyRec := &RecordInfos{}
	rec := RecordInfo{}
	bodyRec.CallInfos = append(bodyRec.CallInfos, rec)
	bodyRec.CallInfos[0].AbonentPhone = "9182222222"
	bodyRec.CallInfos[0].CallDate = time.Now()
	bodyRec.CallInfos[0].CallDirection = "INB"
	bodyRec.CallInfos[0].RecordId = 1
	bodyRec.CallInfos[0].Duration = 101809
	bodyRec.CallInfos[0].ClientPhone = "9060000000"
	bodyRec.CallInfos[0].Provider = "Beeline"
	bodyRec.CallInfos[0].InternalId = 345454
	bodyRec.CallInfos[0].FileSize = 4454
	bodyRec.Count = 1
	httpmock.Activate()
	defer httpmock.Deactivate()
	httpmock.RegisterResponder("PUT", "https://cloudpbx.beeline.ru/api/pub/client/call/record/list",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewXmlResponse(200, bodyRec)
			if err != nil {
				return httpmock.NewStringResponse(500, err.Error()), nil
			}
			return resp, nil
		})
	bodyStr := BuildXMLRequest("INB")
	rs, err := GetRecordsInfoFromServer(bodyStr)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if rs == nil {
		log.Fatalf("Структура с информацией о записях пуста, ожидалось наличие %d записи", num)
	}
	if rs.Count != 1 {
		log.Fatalln("Распознано неверное количество записей")
	}
	if rs.Len() != 1 {
		log.Fatalln("Структура с информацией о записях заполнена неверно")
	}
}
func TestGetWavFilesFromServer(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@/gfk?charset=utf8")
	if err != nil {
		log.Fatalln("Ошибка при подключении к БД" + err.Error())
	}
	defer db.Close()
	testFile, err := os.Open("1.wav")
	if err != nil {
		log.Fatalln("Не удалось открыть тестовый файл записи" + err.Error())
	}
	body, err := ioutil.ReadAll(testFile)
	if err != nil {
		log.Fatalln("Не прочитать тестовый файл записи" + err.Error())
	}
	httpmock.Activate()
	defer httpmock.Deactivate()
	httpmock.RegisterResponder("GET", "https://cloudpbx.beeline.ru/api/pub/client/call/record/file/1",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewBytesResponse(200, body)
			return resp, nil
		})
	recs := RecordInfos{}
	rec := RecordInfo{}
	recs.CallInfos = append(recs.CallInfos, rec)
	recs.CallInfos[0] = RecordInfo{RecordId: int64(1), Status: "failed"}
	var ir IRecordsInfoProvider
	ir = &recs
	err = GetWavFilesFromServer(ir, db)
	if err != nil {
		log.Fatalln(err)
	}
	todayWavFolder := "wav" + string(filepath.Separator) + time.Now().Format("02-01-2006") + string(filepath.Separator)
	_, err = os.Open(todayWavFolder + strconv.FormatInt(recs.CallInfos[0].RecordId, 10) + ".wav")
	if err != nil {
		log.Fatalln(err)
	}
}
func TestSaveRecordInfoToDB(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@/gfk?charset=utf8")
	if err != nil {
		log.Fatalln("Ошибка при подключении к БД" + err.Error())
	}
	defer db.Close()
	recs := RecordInfos{}
	rec := RecordInfo{}
	recs.CallInfos = append(recs.CallInfos, rec)
	recs.CallInfos[0] = RecordInfo{RecordId: int64(1), Status: "failed", Duration: 2323, FileSize: 2344, CallDate: time.Now()}
	err = SaveRecordInfoToDB(&recs.CallInfos[0], db)
	if err != nil {
		log.Fatalln(err.Error())
	}
	query := fmt.Sprintf("SELECT record_id,status,save_date FROM beeline_files WHERE record_id=%d", recs.GetRecordInfo(0).GetId())
	row := db.QueryRow(query)
	testRecordInfo := RecordInfo{}
	row.Scan(&testRecordInfo.RecordId, &testRecordInfo.Status, &testRecordInfo.SaveDate)
	if testRecordInfo.RecordId != recs.GetRecordInfo(0).GetId() {
		if err != nil {
			errMsg := fmt.Sprintf("Информация об ID записи неверно сохранена в базе, ожидалось %d,получено %d", recs.GetRecordInfo(0).GetId(), testRecordInfo.RecordId)
			log.Fatalln(errMsg)
		}
	}
	if testRecordInfo.Status != recs.GetRecordInfo(0).GetStatus() {
		if err != nil {
			errMsg := fmt.Sprintf("Информация о статусе записи неверно сохранена в базе, ожидалось %s,получено %s", recs.GetRecordInfo(0).GetStatus(), testRecordInfo.Status)
			log.Fatalln(errMsg)
		}
	}
	if testRecordInfo.Duration != recs.CallInfos[0].Duration {
		if err != nil {
			errMsg := fmt.Sprintf("Информация о продолжительности записи неверно сохранена в базе, ожидалось %d,получено %d", recs.CallInfos[0].Duration, testRecordInfo.Duration)
			log.Fatalln(errMsg)
		}
	}
	if testRecordInfo.FileSize != recs.CallInfos[0].FileSize {
		if err != nil {
			errMsg := fmt.Sprintf("Информация о размере файла неверно сохранена в базе, ожидалось %d,получено %d", recs.CallInfos[0].FileSize, testRecordInfo.FileSize)
			log.Fatalln(errMsg)
		}
	}
	if testRecordInfo.CallDate != recs.CallInfos[0].CallDate {
		if err != nil {
			errMsg := fmt.Sprintf("Информация о времени разговора неверно сохранена в базе, ожидалось %d,получено %d", recs.CallInfos[0].CallDate, testRecordInfo.CallDate)
			log.Fatalln(errMsg)
		}
	}
}
func TestConvertWavToMp3Files(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@/gfk?charset=utf8")
	if err != nil {
		log.Fatalln("Ошибка при подключении к БД" + err.Error())
	}
	recs := RecordInfos{}
	rec := RecordInfo{}
	recs.CallInfos = append(recs.CallInfos, rec)
	recs.CallInfos[0] = RecordInfo{RecordId: int64(1), Status: "failed", Duration: 2323, FileSize: 2344, CallDate: time.Now()}

	defer db.Close()
	err = ConvertWavToMp3Files(&recs, db)
	if err != nil {
		log.Fatalln("Ошибка при подключении к БД" + err.Error())
	}
	todayMp3Folder := "mp3" + string(filepath.Separator) + time.Now().Format("02-01-2006") + string(filepath.Separator)
	_, err = os.Open(todayMp3Folder + strconv.FormatInt(recs.CallInfos[0].RecordId, 10) + ".mp3")
	if err != nil {
		log.Fatalln(err)
	}
}
