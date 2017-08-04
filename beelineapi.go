package beelineapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	"strconv"
)

const (
	// Статусы агента call-центра
	ONLINE  = 0
	OFFLINE = 1
	BREAK   = 2
	// CONTENTTYPE Тип ответа
	CONTENTTYPE string = "application/json"
	// Статус записи разговоров для абонента
	OFF = 0
	ON  = 1
)

// APIClient структура для хранения информации об абоненте
type APIClient struct {
	Token      string
	Params     []string
	Provider   string
	BaseApiUrl string
}

//APIError Структура для хранения ошибок от сервера
type APIError struct {
	ErrorCode   string `json:"errorCode"`   // Код ошибки
	Description string `json:"description"` // Текст ошибки
}

//WrapError Тип хранения ошибок
type WrapError struct {
	Msg string
}

func (d WrapError) Error() string {
	return d.Msg
}

// Abonent структура для хранения информации об абоненте
type Abonent struct {
	UserId     string `json:"userId"`
	Phone      string `json:"phone"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Extension  string `json:"extension"`
}

// Abonents структура для хранения информации об абонентах
type Abonents []Abonent

// CfsStatusResponse
type CfsStatusResponse struct {
	IsCfsServiceEnabled bool
	RuleList            []CfsRule
}

// CfsRule
type CfsRule struct {
	Id             int      `json:""`
	Name           string   `json:""`
	ForwardToPhone string   `json:""`
	Schedule       int      `json:""` //[ROUND_THE_CLOCK (Круглосуточно), WORKING_TIME (Рабочее время), NON_WORKING_TIME_AND_HOLIDAYS (Нерабочие часы и выходные)]
	PhoneList      []string `json:""`
}

// CfsRuleUpdate Запрос для добавления правила
type CfsRuleUpdate struct {
	Name           string
	ForwardToPhone string
	Schedule       int // [ROUND_THE_CLOCK (Круглосуточно), WORKING_TIME (Рабочее время), NON_WORKING_TIME_AND_HOLIDAYS (Нерабочие часы и выходные)]
	PhoneList      []string
}

// BasicRedirect Номера для переадресации
type BasicRedirect struct {
	ForwardAllCallsPhone    string `json:"forwardAllCallsPhone"`
	ForwardBusyPhone        string `json:"forwardBusyPhone"`
	ForwardUnavailablePhone string `json:"forwardUnavailablePhone"`
	ForwardNotAnswerPhone   string `json:"forwardNotAnswerPhone"`   //Номер, на который будет выполнена переадресация если номер не отвечает
	ForwardNotAnswerTimeout int    `json:"forwardNotAnswerTimeout"` //Колличество гудков, которые необходимо подождать для ответа номера
}

// BasicRedirectResponse Возвращаемое значение:
type BasicRedirectResponse struct {
	Status  int           `json:"status"`  // Статус переадресации = [ON (Переадресация включена), OFF (Переадресация выключена)]
	Forward BasicRedirect `json:"forward"` // Номера для переадресации
}

// BwlStatusResponse
type BwlStatusResponse struct {
	Status    int       `json:"status"` // [BLACK_LIST_ON (Не принимать звонки с указанных в списке правил номеров), WHITE_LIST_ON (Принимать звонки только с указанных в списке правил номеров), OFF (Услуга отключена)]
	BlackList []BwlRule `json:"blackList"`
	WhiteList []BwlRule `json:"whiteList"`
}

// BwlRule
type BwlRule struct {
	Id        int      `json:"id"`
	Name      string   `json:"name"`
	Schedule  int      `json:"schedule"` // [ROUND_THE_CLOCK (Круглосуточно), WORKING_TIME (Рабочее время), NON_WORKING_TIME_AND_HOLIDAYS (Нерабочие часы и выходные)]
	PhoneList []string `json:"phoneList"`
}

// BwlRuleAdd
type BwlRuleAdd struct {
	Type int           `json:"type"` // [BLACK_LIST (Не принимать звонки с указанных в списке правил номеров), WHITE_LIST (Принимать звонки только с указанных в списке правил номеров)]
	Rule BwlRuleUpdate `json:"rule"`
}

// BwlRuleUpdate Запрос для обновления правила
type BwlRuleUpdate struct {
	Name      string   `json:"name"`
	Schedule  int      `json:"schedule"` // [ROUND_THE_CLOCK (Круглосуточно), WORKING_TIME (Рабочее время), NON_WORKING_TIME_AND_HOLIDAYS (Нерабочие часы и выходные)]
	PhoneList []string `json:"phoneList"`
}
type NumberInfo struct {
	NumberId string `json:"numberId"` // Идентификатор входящего номера
	Phone    string `json:"phone"`    //Номер телефона
}
type SubscriptionRequest struct {
	Pattern          string `json:"pattern"`          //Идентификатор, входящий или добавочный номер абонента или номера
	Expires          int    `json:"expires"`          //Длительность подписки
	SubscriptionType int    `json:"subscriptionType"` // Тип подписки = [BASIC_CALL (Базовая информация о вызове), ADVANCED_CALL (Расширеная информация о вызове)]
	Url              string `json:"url"`
}
type SubscriptionResult struct {
	SubscriptionId string `json:"subscriptionId"` //Идентификатор подписки
	Expires        int    `json:"expires"`        //Длительность подписки
}
type SubscriptionInfo struct {
	SubscriptionId   string `json:"subscriptionId"`   //Идентификатор подписки
	TargetType       int    `json:"targetType"`       //Тип объекта, для которого сформирована подписка = [GROUP (События всей группы), ABONENT (События абонента), NUMBER (События номера)]
	TargetId         string `json:"targetId"`         //Идентификатор объекта, для которого сформирована подписка
	SubscriptionType int    `json:"subscriptionType"` //Тип подписки = [BASIC_CALL (Базовая информация о вызове), ADVANCED_CALL (Расширеная информация о вызове)]
	Expires          int    `json:"expires "`         //Длительность подписки
	Url              string `json:"url"`              //URL приложения
}
type IcrNumbersResult struct {
	PhoneNumber string            `json:"phoneNumber"` //Номер телефона
	Status      int               `json:"status"`      //Результат выполнения операции = [SUCCESS (Успешно), FAULT (Ошибка)]
	Error       IcrOperationError `json:"error"`       //Описание ошибки
}
type IcrOperationError struct {
	ErrorCode   string `json:"errorCode"`   //Код ошибки
	Description string `json:"description"` //Сообщение об ошибке
}

// IcrRouteRule структура хранения правила переадресации
type IcrRouteRule struct {
	InboundNumber string `json:"inboundNumber"` //Входящий номер клиента
	Extension     string `json:"extension"`     //Внутренний номер
}

// IcrRouteResult структура хранения статуса удаления правил переадресации
type IcrRouteResult struct {
	Rule   IcrRouteRule      `json:"rule"`   //Правило переадресации
	Status int               `json:"status"` //Результат выполнения операции = [SUCCESS (Успешно), FAULT (Ошибка)]
	Error  IcrOperationError `json:"error"`  //Описание ошибки
}

type UnixNano struct {
	time.Time
}

// CallRecord структура хранения подробной информации об отдельной записи
type CallRecord struct {
	Id         int           `json:"id,string"`  //Идентификатор записи
	ExternalId string        `json:"externalId"` //Внешний идентификатор записи
	Phone      string        `json:"phone"`      //Мобильный номер абонента
	Direction  string        `json:"direction"`  //Тип вызова = [INBOUND (Входящий вызов), OUTBOUND (Исходящий вызов)]
	Date       UnixNano      `json:"date"`       //Дата и время разговора
	Duration   int           `json:"duration"`   //Длительность разговора в миллисекундах
	FileSize   int           `json:"fileSize"`   //Размер файла записи разговора
	Comment    string        `json:"comment"`    //Комментарий к записи разговора
	Abonent    Abonent       `json:"abonent"`    //Абонент
}

func (t *UnixNano) MarshalJSON() ([]byte, error) {
	ts := t.Time.UnixNano()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

func (t *UnixNano) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}

	t.Time = time.Unix(int64(ts) / int64(time.Microsecond), int64(ts) % int64(time.Microsecond))

	return nil
}

//  ------------------------------------- Операции с записями разговоров  -------------------------------------

// GetRecord Возвращает список записей разговоров
// id - Идентификатор записи
func (c APIClient) GetRecord(recordId int) (CallRecord, error) {
	url := fmt.Sprintf("%sv2/records/%d", c.BaseApiUrl, recordId)

	recs := CallRecord{}
	body, err := createRequest("GET", url, c.Token, "")

	if err != nil {
		return recs, err
	}
	if err := json.Unmarshal(body, &recs); err != nil {
		return recs, err
	}
	return recs, nil
}

// GetRecords Возвращает описание записи разговора
func (c APIClient) GetRecords() ([]CallRecord, error) {
	url := c.BaseApiUrl + "records"

	recs := []CallRecord{}
	body, err := createRequest("GET", url, c.Token, "")

	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &recs); err != nil {
		return nil, err
	}
	return recs, nil
}

// DeleteRecord Удаляет запись разговора
// id - Идентификатор записи разговора
func (c APIClient) DeleteRecord(recordId int) error {
	url := fmt.Sprintf("%sv2/records/%d", c.BaseApiUrl, recordId)

	_, err := createRequest("DELETE", url, c.Token, "")
	if err != nil {
		return WrapError{Msg: "Ошибка при удалении записи с сервера Билайн. " + err.Error()}
	}
	return nil
}

// GetRecordFile Возвращает файл записи разговора
// id - Идентификатор разговора
func (c APIClient) GetRecordFile(recordId int) (io.Reader, error) {
	var r io.Reader
	url := fmt.Sprintf("%sv2/records/%d/download", c.BaseApiUrl, recordId)
	body, err := createRequest("GET", url, c.Token, "")
	if err != nil {
		return nil, WrapError{Msg: "Ошибка при подготовке запроса на получение информации о записях разговоров. " + err.Error()}
	}
	r = bytes.NewReader(body)
	return r, nil
}

// createRequest Функция отправки запроса
// method - метод HTTP запроса
// url - адрес
// token - токен авторизации
// body - тело запроса
func createRequest(method string, url string, token string, b string) ([]byte, error) {
	body := strings.NewReader(b)
	recordReq, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, WrapError{Msg: "Ошибка при подготовке запроса к серверу Beeline. " + err.Error()}
	}

	// Устанавливаем HTTP заголовок билайновский для ключа безопасности
	recordReq.Header.Set("X-MPBX-API-AUTH-TOKEN", token)
	// Установка времени ожидания ответа от сервера равной 10 секундам
	timeout := time.Duration(60 * time.Second)
	cl := &http.Client{Timeout: timeout}
	recordReq.Close = true

	resp, err := cl.Do(recordReq)
	if err != nil {
		return nil, WrapError{Msg: "Ошибка при отправке запроса к серверу Beeline. " + err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, WrapError{Msg: fmt.Sprintf(
			"Ошибка при запросе к серверу Beeline. Получен HTTP код ответа %d. %s",
			resp.StatusCode,
			resp.Status,
		)}
	}

	responseBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, WrapError{Msg: "Ошибка при чтении ответа после отправке запроса к серверу Beeline. " + err.Error()}
	}

	return responseBody, nil
}

func fireError(err error, msg string) {
	if err != nil {
		log.Fatalln(msg + err.Error())
	}
}

func NewApiClient(token string) (c APIClient) {
	c = APIClient{}
	c.Token = token
	c.Provider = "Beeline"
	c.BaseApiUrl = "https://cloudpbx.beeline.ru/apis/portal/"
	return c
}
