package beelapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
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

//BeeAPIError Тип хранения ошибок
type BeeAPIError struct {
	Msg string
}

func (d BeeAPIError) Error() string {
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

// Abonents структура для хранения информации об абоненте
type Abonents struct {
	Abnts []Abonent
}

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

// CallRecord структура хранения подробной информации об отдельной записи
type CallRecord struct {
	Id         string    `json:"id"`         //Идентификатор записи
	ExternalId string    `json:"externalId"` //Внешний идентификатор записи
	Phone      string    `json:"phone"`      //Мобильный номер абонента
	Direction  int       `json:"direction"`  //Тип вызова = [INBOUND (Входящий вызов), OUTBOUND (Исходящий вызов)]
	Date       time.Time `json:"date"`       //Дата и время разговора
	Duration   int       `json:"duration"`   //Длительность разговора в миллисекундах
	FileSize   int       `json:"fileSize"`   //Размер файла записи разговора
	Comment    string    `json:"comment"`    //Комментарий к записи разговора
	Abonent    Abonent   `json:"abonent"`    //Абонент
}

// APIClient структура для хранения информации об абоненте
type APIClient struct {
	Token    string
	Params   []string
	Provider string
}

// TimeRange структура хранения начальной и конечной дат, за которые запрашиваются записи разговоров
type TimeRange struct {
	StartStamp time.Time
	EndStamp   time.Time
}

var cfg APIClient

//  ------------------------------------- Операции с абонентами -------------------------------------

//  ------------------------------------- Простая переадресация вызовов -------------------------------------
// GetAbonent Возвращает список всех абонентов
func (c APIClient) GetAbonents() Abonents {

}

// GetAbonents Ищет абонента по идентификатору, мобильному или добавочному номеру
// id - Идентификатор, мобильный или добавочный номер абонента
func (c APIClient) GetAbonent(id string) Abonent {

}

// GetAgentStatus Возвращает статус агента call-центра
// id - Идентификатор, мобильный или добавочный номер абонента
func (c APIClient) GetAgentStatus(id string) int {

}

// SetAgentStatus Устанавливает статус агента call-центра
// id - Идентификатор, мобильный или добавочный номер абонента
// newStatus - Новый статус агента
func (c APIClient) SetAgentStatus(id string, newStatus string) {

}

// GetRecordingStatus Возвращает статус записи разговоров для абонента
// id - Идентификатор, мобильный или добавочный номер абонента
func (c APIClient) GetRecordingStatus(id string) int {

}

// TurnOnRecording Включает запись разговоров для абонента
// id - Идентификатор, мобильный или добавочный номер абонента
func (c APIClient) TurnOnRecording(id string) error {

}

// TurnOffRecording Отключает запись разговоров для абонента
// id - Идентификатор, мобильный или добавочный номер абонента
func (c APIClient) TurnOffRecording(id string) error {

}

//DoCall Совершает звонок от имени абонента
// id - Идентификатор, мобильный или добавочный номер абонента
// telNumber -Номер телефона - 10 цифр
func (c APIClient) DoCall(id string, telNumber string) string {

}

// TurnOnNumberToAbonent Подключает дополнительный номер абоненту
// id - Идентификатор, мобильный или добавочный номер абонента
// telNumber -Подключаемый номер телефона - 10 цифр
// schedule - Расписание перенаправления на номер
func (c APIClient) TurnOnNumberToAbonent(id string, telNumber string, schedule int) error {

}

// TurnOffNumberToAbonent Отключает дополнительный номер абонента
// id - Идентификатор, мобильный или добавочный номер абонента
func (c APIClient) TurnOffNumberToAbonent(id string) error {

}

//  ------------------------------------- Простая переадресация вызовов -------------------------------------

// GetBasicRedirectStatus Возвращает статус базовой переадресации
// id - Идентификатор, мобильный или добавочный номер абонента
func (c APIClient) GetBasicRedirectStatus(id string) int {

}

// TurnOnBasicRedirect Включает базовую переадресацию
// id - Идентификатор, мобильный или добавочный номер абонента
// br - Номера для переадресации
func (c APIClient) TurnOnBasicRedirect(id string, br BasicRedirect) error {

}

// TurnOffBasicRedirect Отключает базовую переадресацию
// id - Идентификатор, мобильный или добавочный номер абонента
func TurnOffBasicRedirect(id string) error {

}

//  ------------------------------------- Выборочная переадресация вызовов -------------------------------------

// Возвращает список правил выборочной переадресации
// id - Идентификатор, мобильный или добавочный номер абонента
func (c APIClient) GetSelectiveCallRules(id string) (CfsStatusResponse, error) {

}

// Добавляет правило для выборочной переадресации
// id - Идентификатор, мобильный или добавочный номер абонента
// rule -Запрос для добавления правила
func (c APIClient) AddSelectiveCallRule(id string, rule CfsRuleUpdate) error {

}

// Включает выборочную переадресацию
// id - Идентификатор, мобильный или добавочный номер абонента
func (c APIClient) TurnOnSelectiveRedirect(id string) error {

}

// UpdateSelectiveCallRule Обновляет правило
// id - Идентификатор, мобильный или добавочный номер абонента
// ruleID -Идентификатор правила
// rule - Запрос для обновления правила
func (c APIClient) UpdateSelectiveCallRule(id string, ruleID int, rule CfsRuleUpdate) error {

}

// TurnOffSelectiveRedirect Отключает выборочную переадресацию
// id - Идентификатор, мобильный или добавочный номер абонента
func (c APIClient) TurnOffSelectiveRedirect(id string) error {

}

// DeleteSelectiveRedirect Удаляет правило
// id - Идентификатор, мобильный или добавочный номер абонента
// ruleID - Идентификатор правила
func (c APIClient) DeleteSelectiveCallRule(id string, ruleID int) error {

}

//  ------------------------------------- Выборочный прием звонков -------------------------------------

// Статус и список правил для выборочного приема звонков
// id - Идентификатор, мобильный или добавочный номер абонента
func (c APIClient) IncCallRules(id string) (BwlStatusResponse, error) {

}

// AddIncCallRule Добавляет правило для выборочного приема звонков
// id - Идентификатор, мобильный или добавочный номер абонента
// ruleUpdate - Запрос для добавления правила
func (c APIClient) AddIncCallRule(id string, rule BwlRuleAdd, ruleUpdate BwlRuleUpdate) int {

}

// TurnOnSelectiveCallReceive Включает выборочный прием звонков
// id - Идентификатор, мобильный или добавочный номер абонента
// t - Тип правила
func (c APIClient) TurnOnSelectiveCallReceive(id string, t int) error {
}

// UpdateSelectiveReceiveRule Обновляет правило для выборочного приема звонков
// id - Идентификатор, мобильный или добавочный номер абонента
// ruleID - Идентификатор правила
// ruleUpdate -Запрос для обновления правила
func (c APIClient) UpdateSelectiveReceiveRule(id string, ruleID int, ruleUpdate BwlRuleUpdate) error {

}

// TurnOffSelectiveReceiveRule Отключает выборочный прием звонков
// id - Идентификатор, мобильный или добавочный номер абонента
func (c APIClient) TurnOffSelectiveReceiveRule(id string) error {

}

// DeleteSelectiveReceiveRule Удаляет правило для выборочного приема звонков
// id - Идентификатор, мобильный или добавочный номер абонента
// ruleID - Идентификатор правила
func DeleteSelectiveReceiveRule(id string, ruleId int) error {

}

//  ------------------------------------- Операции с записями разговоров  -------------------------------------

// GetRecords Записи разговоров передаются по порядку начиная со следующей после переданного
// ID или с первой записи, если ID не передан. За один запрос передаётся не более чем 100 записей.
// id - Начальный ID записи
func (c APIClient) GetRecords(id string) ([]CallRecord, error) {
	url = "https://cloudpbx.beeline.ru/apis/portal/records"
	recs := []CallRecord{}
	recs = json.Unmarshal(createRequest("GET", url, ""))
	fmt.Println(recs[0].Phone)
}

// DeleteRecord Удаляет запись разговора по уникальному идентификатору записи recordId.
// id - Идентификатор записи разговора
func (c APIClient) DeleteRecord(id string) error {

}

// GetRecordInfo Возвращает запись разговора по уникальному идентификатору записи recordId.
// id - Идентификатор записи разговора
func (c APIClient) GetRecordInfo(id string) (CallRecord, error) {

}

// GetRecordInfoFromEvent Возвращает запись разговора по ID разговора из события и ID пользователя из того же события.
// id - Идентификатор разговора из события
// userId - Идентификатор пользователя из события
func (c APIClient) GetRecordInfoFromEvent(id string, userId string) (CallRecord, error) {

}

// GetRecordFile Возвращает файл записи разговора по уникальному идентификатору записи recordId
// id - Идентификатор разговора из события

func (c APIClient) GetRecordFile(id string) (Reader, error) {

}

// GetRecordFileFromEvent Возвращает запись разговора по ID разговора  из события и ID пользователя из того же события.
// id - Идентификатор разговора из события
// userId - Идентификатор пользователя из события

func (c APIClient) GetRecordFileFromEvent(id string, userId string) (Reader, error) {

}

//  ------------------------------------- Операции со входящими номерами  -------------------------------------

// GetAllIncNumbers Возвращает список всех входящих номеров
// id - Идентификатор разговора из события
// userId - Идентификатор пользователя из события

func (c APIClient) GetAllIncNumbers() ([]NumberInfo, error) {

}

// FindIncNumberById Ищет входящий номер по идентификатору, номеру или добавочному номеру
// id - Идентификатор, номер или добавочный номер

func (c APIClient) FindIncNumberById(id string) (NumberInfo, error) {

}

//  ------------------------------------- Подписка на Xsi-Events  -------------------------------------

// XSIEventSubscribtion Формирует подписку на Xsi-Events
// Подписка может быть использована для интеграции со сторонними системами, которым необходим контроль над звонками абонентов облачной АТС в реальном времени.
// API использует механизм подписки на события, ассоциированные с тем или иным абонентом, номером или всем клиентом.
// Например, Абонент облачной АТС принимает вызов, сторонняя CRM система получает обновления о текущем статусе вызова (ringing, established, completed).
// req - Запрос для подписки на события

func (c APIClient) XSIEventSubscription(reg SubscriptionRequest) (SubscriptionResult, error) {

}

// GetXSIEventSubscriptionInfo Возвращает информацию о подписке на Xsi-Events
// id - Идентификатор подписки

func (c APIClient) GetXSIEventSubscriptionInfo(id string) (SubscriptionInfo, error) {

}

// TurnOffXSIEventSubscription Отключает подписку на Xsi-Events
// id - Идентификатор отключаемой подписки

func (c APIClient) TurnOffXSIEventSubscription(id string) error {

}

//  ------------------------------------- Индивидуальная переадресация  -------------------------------------

// GetIncNumWithRedirect Возвращает список входящих номеров, для которых включена переадресация

func (c APIClient) GetIncNumWithRedirect() ([]NumberInfo, error) {

}

// TurnOnCustomIncNumRedirect Включает индивидуальную переадресацию для входящих номеров
//  numberList - Список входящих номеров, для которых должна быть включена переадресация
func (c APIClient) TurnOnCustomIncNumRedirect(numberList []string) ([]IcrNumbersResult, error) {

}

// TurnOffCustomIncNumRedirect Отключает индивидуальную переадресацию для входящих номеров
//  numberList - Список входящих номеров, для которых должна быть отключена переадресация
func (c APIClient) TurnOffCustomIncNumRedirect(numberList []string) ([]IcrNumbersResult, error) {

}

// GetRedirectRulesList Возвращает список правил переадресации
func (c APIClient) GetRedirectRulesList() ([]IcrRouteRule, error) {

}

// DeleteRedirectRulesList Удаляет список правил переадресации
// rules - Список правил переадресации
func (c APIClient) DeleteRedirectRulesList(rules []IcrRouteRule) ([]IcrRouteResult, error) {

}

// ReplaceRedirectRulesList Замещает правила переадресации
// rules - Список правил переадресации
func (c APIClient) ReplaceRedirectRulesList(rules []IcrRouteRule) ([]IcrRouteResult, error) {

}

// UnionRedirectRulesList Объединяет существующие правила переадресации с переданным списком правил.
// rules - Список правил переадресации
func (c APIClient) UnionRedirectRulesList(rules []IcrRouteRule) ([]IcrRouteResult, error) {

}

// createRequest Функция отправки запроса
// reqType - тип HTTP запроса
// url - адрес
// body - тело запроса
func createRequest(reqType string, url string, b string) ([]byte, error) {
	body := strings.NewReader(b)
	recordReq, err := http.NewRequest(reqType, url, body)
	if err != nil {
		return nil, BeeAPIError{Msg: "Ошибка при подготовке запроса к серверу Beeline на получение файлов записей" + err.Error()}
	}
	recordReq.Header.Set("X-MPBX-API-AUTH-TOKEN", APIClient.Token)
	q := recordReq.URL.Query
	for i := 0; i < len(params); i++ {
		q.add()
	}
	cl := &http.Client{}
	resp, err := cl.Do(recordReq)
	if err != nil {
		return nil, BeeAPIError{Msg: "Ошибка при отправке запроса к серверу Beeline на получение файлов записей" + err.Error()}
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, BeeAPIError{Msg: "Ошибка при чтении ответа после отправке запроса к серверу Beeline на получение файлов записей" + err.Error()}
	}
	return body, nil
}

// createUrlWithQuery Функция формирования url по шаблону, включая строку с параметрами
// url - адрес
// params - параметры запроса
func createUrlWithQuery(url string, params []string) string {
	if len(params) == 0 {
		return url
	}
	r := http.NewRequest
	q := r.URL.Query
	var re = regexp.MustCompile(`{[^\s{}]*}`)

	for i := 0; i < len(params); i++ {
		s := re.FindStringSubmatchIndex(url, -1)
		url = strings.Replace(url, url[s[0]:s[1]], params[i], -1)

	}
	return url
}
