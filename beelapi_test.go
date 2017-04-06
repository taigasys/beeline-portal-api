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
	"gopkg.in/yaml.v2"
)

type Config struct {
	PeriodInHours int    `yaml:"period_in_hours"`
	Provider      string `yaml:"provider"`
	DBTable       string `yaml:"db_table"`
	DBName        string `yaml:"db_name"`
	DBHost        string `yaml:"db_host"`
	DBUser        string `yaml:"db_user"`
	DBPassword    string `yaml:"db_psw"`
	RecordListUrl string `yaml:"record_list_url"`
	RecordFileUrl string `yaml:"record_file_url"`
}

var todayMp3Folder string
var todayWavFolder string
var cfg = Config{}
var dbinfo string

func init() {
	bts, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalln("Не удалось подключить файл конфигурации" + err.Error())
	}
	err = yaml.Unmarshal(bts, &cfg)
	if err != nil {
		log.Fatalln("Не удалось подключить файл конфигурации" + err.Error())
	}
	todayMp3Folder = "mp3" + string(filepath.Separator) + time.Now().Format("02-01-2006") + string(filepath.Separator)
	todayWavFolder = "wav" + string(filepath.Separator) + time.Now().Format("02-01-2006") + string(filepath.Separator)
	dbinfo = "user=" + cfg.DBUser + " password=" + cfg.DBPassword + " dbname=" + cfg.DBName + " host=" + cfg.DBHost + " sslmode=disable"
	db, err := sql.Open("postgres", dbinfo)
	defer db.Close()
	if err != nil && db == nil {
		log.Fatalln("Не удалось подключиться к серверу БД" + err.Error())
	}
	_, _ = db.Exec("CREATE TYPE  calldirection AS ENUM ('INB','OUT');CREATE TYPE callstatus AS ENUM ('saved','failed');")
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS \"" + cfg.DBTable + "\"" +
		"(" +
		"id SERIAL, " +
		"record_id bigint NOT NULL," +
		"internal_id bigint DEFAULT NULL," +
		"abonent char(11) NOT NULL," +
		"phone char(11) NOT NULL," +
		"call_direction calldirection NOT NULL, " +
		"call_date timestamp with time zone NOT NULL, " +
		"duration integer NOT NULL," +
		"file_size bigint NOT NULL," +
		"status callstatus DEFAULT 'failed'," +
		"save_date timestamp with time zone DEFAULT now(), " +
		"provider character varying(32) DEFAULT 'Beeline'," +
		"CONSTRAINT pk_record_id PRIMARY KEY (id)," +
		"CONSTRAINT un_record_id UNIQUE (record_id)" +
		")" +
		"WITH (" +
		"OIDS=FALSE);" +
		"CREATE INDEX IF NOT EXISTS ind_record_id ON " + cfg.DBTable + " USING btree(record_id);")
	if err != nil {
		log.Fatalln("Не удалось создать таблицу в БД" + err.Error())
	}
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
	httpmock.RegisterResponder("PUT", cfg.RecordListUrl,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewXmlResponse(200, bodyRec)
			if err != nil {
				return httpmock.NewStringResponse(500, err.Error()), nil
			}
			return resp, nil
		})
	timeRange := TimeRange{StartStamp: time.Now().Add(-time.Duration(10) * 60 * time.Minute), EndStamp: time.Now()}
	bodyStr := BuildXMLRequest("INB", timeRange)
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
func TestGetWavFileFromServer(t *testing.T) {
	db, err := sql.Open("postgres", dbinfo)
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
	httpmock.RegisterResponder("GET", cfg.RecordFileUrl+"1",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewBytesResponse(200, body)
			return resp, nil
		})
	rec := RecordInfo{}
	var ir IRecordInfoProvider
	rec.RecordId = int64(1)
	rec.Status = "false"
	rec.Duration = 101809
	rec.CallDate = time.Now()
	rec.FileSize = 2344
	ir = &rec
	isAlreadyHandled, err := GetWavFileFromServer(ir, todayWavFolder, db)
	if err != nil {
		log.Fatalln(err)
	}
	if isAlreadyHandled {
		return
	}
	_, err = os.Open(todayWavFolder + strconv.FormatInt(rec.RecordId, 10) + ".wav")
	if err != nil {
		log.Fatalln(err)
	}
}
func TestSaveRecordInfoToDB(t *testing.T) {
	db, err := sql.Open("postgres", dbinfo)
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
	query := fmt.Sprintf("SELECT record_id,status,save_date FROM "+cfg.DBTable+" WHERE record_id=%d", recs.GetRecordInfo(0).GetId())
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
func TestConvertWavToMp3File(t *testing.T) {
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatalln("Ошибка при подключении к БД" + err.Error())
	}
	defer db.Close()
	rec := RecordInfo{}

	rec.RecordId = int64(1)
	rec.Status = "failed"
	rec.Duration = 2323
	rec.FileSize = 2344
	rec.CallDate = time.Now()

	defer db.Close()
	err = ConvertWavToMp3File(&rec, todayWavFolder, todayMp3Folder)
	if err != nil {
		log.Fatalln(err.Error())
	}
	todayMp3Folder := "mp3" + string(filepath.Separator) + time.Now().Format("02-01-2006") + string(filepath.Separator)
	_, err = os.Open(todayMp3Folder + strconv.FormatInt(rec.RecordId, 10) + ".mp3")
	if err != nil {
		log.Fatalln(err)
	}
}
