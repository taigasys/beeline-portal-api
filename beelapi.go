package beelapi

import "time"

const (
	// Статусы агента call-центра
	ONLINE = iota
	OFFLINE
	BREAK
	// CONTENTTYPE Тип ответа
	CONTENTTYPE string = "application/json"
	// Статус записи разговоров для абонента
	ON
	OFF
)

// Abonent структура для хранения информации об абоненте
type Abonent struct {
	UserId     string `json:""`
	Phone      string `json:""`
	FirstName  string `json:""`
	LastName   string `json:""`
	Email      string `json:""`
	Department string `json:""`
	Extension  string `json:""`
}

// Abonents структура для хранения информации об абоненте
type Abonents struct {
	Abnts []Abonent `json:""`
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
	ForwardAllCallsPhone    string `json:""`
	ForwardBusyPhone        string `json:""`
	ForwardUnavailablePhone string `json:""`
	ForwardNotAnswerPhone   string `json:""` //Номер, на который будет выполнена переадресация если номер не отвечает
	ForwardNotAnswerTimeout int    `json:""` //Колличество гудков, которые необходимо подождать для ответа номера
}

// BRResponse Возвращаемое значение:
type BRResponse struct {
	Status  int           `json:""` // Статус переадресации = [ON (Переадресация включена), OFF (Переадресация выключена)]
	Forward BasicRedirect `json:""` // Номера для переадресации
}

// BwlStatusResponse
type BwlStatusResponse struct {
	Status    int       `json:""` // [BLACK_LIST_ON (Не принимать звонки с указанных в списке правил номеров), WHITE_LIST_ON (Принимать звонки только с указанных в списке правил номеров), OFF (Услуга отключена)]
	BlackList []BwlRule `json:""`
	WhiteList []BwlRule `json:""`
}

// BwlRule
type BwlRule struct {
	Id        int      `json:""`
	Name      string   `json:""`
	Schedule  int      `json:""` // [ROUND_THE_CLOCK (Круглосуточно), WORKING_TIME (Рабочее время), NON_WORKING_TIME_AND_HOLIDAYS (Нерабочие часы и выходные)]
	PhoneList []string `json:""`
}

// BwlRuleAdd
type BwlRuleAdd struct {
	Type int           `json:""` // [BLACK_LIST (Не принимать звонки с указанных в списке правил номеров), WHITE_LIST (Принимать звонки только с указанных в списке правил номеров)]
	Rule BwlRuleUpdate `json:""`
}

// BwlRuleUpdate Запрос для обновления правила
type BwlRuleUpdate struct {
	Name      string   `json:""`
	Schedule  int      `json:""` // [ROUND_THE_CLOCK (Круглосуточно), WORKING_TIME (Рабочее время), NON_WORKING_TIME_AND_HOLIDAYS (Нерабочие часы и выходные)]
	PhoneList []string `json:""`
}
type NumberInfo struct {
	NumberId string `json:""` // Идентификатор входящего номера
	Phone    string `json:""` //Номер телефона
}
type SubscriptionRequest struct {
	Pattern          string `json:""` //Идентификатор, входящий или добавочный номер абонента или номера
	Expires          int    `json:""` //Длительность подписки
	SubscriptionType int    `json:""` // Тип подписки = [BASIC_CALL (Базовая информация о вызове), ADVANCED_CALL (Расширеная информация о вызове)]
	Url              string `json:""`
}
type SubscriptionResult struct {
	SubscriptionId string `json:""` //Идентификатор подписки
	Expires        int    `json:""` //Длительность подписки
}
type SubscriptionInfo struct {
	SubscriptionId   string `json:""` //Идентификатор подписки
	TargetType       int    `json:""` //Тип объекта, для которого сформирована подписка = [GROUP (События всей группы), ABONENT (События абонента), NUMBER (События номера)]
	TargetId         string `json:""` //Идентификатор объекта, для которого сформирована подписка
	SubscriptionType int    `json:""` //Тип подписки = [BASIC_CALL (Базовая информация о вызове), ADVANCED_CALL (Расширеная информация о вызове)]
	Expires          int    `json:""` //Длительность подписки
	Url              string `json:""` //URL приложения
}

// CallRecord структура хранения подробной информации об отдельной записи
type CallRecord struct {
	Id         string    //Идентификатор записи
	ExternalId string    //Внешний идентификатор записи
	Phone      string    //Мобильный номер абонента
	Direction  int       //Тип вызова = [INBOUND (Входящий вызов), OUTBOUND (Исходящий вызов)]
	Date       time.Time //Дата и время разговора
	Duration   int       //Длительность разговора в миллисекундах
	FileSize   int       //Размер файла записи разговора
	Comment    string    //Комментарий к записи разговора
	Abonent    Abonent   //Абонент
}

// APIClient структура для хранения информации об абоненте
type APIClientSettings struct {
	Username      string
	Pwd           string
	PeriodInHours int
	RecordListUrl string
	RecordFileUrl string
	Provider      string
}

// TimeRange структура хранения начальной и конечной дат, за которые запрашиваются записи разговоров
type TimeRange struct {
	StartStamp time.Time
	EndStamp   time.Time
}

var cfg APIClientSettings

//BeeAPIError Тип хранения ошибок
type BeeAPIError struct {
	Msg string
}

func (d BeeAPIError) Error() string {
	return d.Msg
}

//  ------------------------------------- Операции с абонентами -------------------------------------
//  ------------------------------------- Простая переадресация вызовов -------------------------------------
// GetAbonent Возвращает список всех абонентов
func GetAbonents() Abonents {

}

// GetAbonents Ищет абонента по идентификатору, мобильному или добавочному номеру
// id - Идентификатор, мобильный или добавочный номер абонента
func GetAbonent(id string) Abonent {

}

// GetAgentStatus Возвращает статус агента call-центра
// id - Идентификатор, мобильный или добавочный номер абонента
func GetAgentStatus(id string) int {

}

// SetAgentStatus Устанавливает статус агента call-центра
// id - Идентификатор, мобильный или добавочный номер абонента
// newStatus - Новый статус агента
func SetAgentStatus(id string, newStatus string) {

}

// GetRecordingStatus Возвращает статус записи разговоров для абонента
// id - Идентификатор, мобильный или добавочный номер абонента
func GetRecordingStatus(id string) int {

}

// TurnOnRecording Включает запись разговоров для абонента
// id - Идентификатор, мобильный или добавочный номер абонента
func TurnOnRecording(id string) error {

}

// TurnOffRecording Отключает запись разговоров для абонента
// id - Идентификатор, мобильный или добавочный номер абонента
func TurnOffRecording(id string) error {

}

//DoCall Совершает звонок от имени абонента
// id - Идентификатор, мобильный или добавочный номер абонента
// telNumber -Номер телефона - 10 цифр
func DoCall(id string, telNumber string) string {

}

// TurnOnNumberToAbonent Подключает дополнительный номер абоненту
// id - Идентификатор, мобильный или добавочный номер абонента
// telNumber -Подключаемый номер телефона - 10 цифр
// schedule - Расписание перенаправления на номер
func TurnOnNumberToAbonent(id string, telNumber string, schedule int) error {

}

// TurnOffNumberToAbonent Отключает дополнительный номер абонента
// id - Идентификатор, мобильный или добавочный номер абонента
func TurnOffNumberToAbonent(id string) error {

}

//  ------------------------------------- Простая переадресация вызовов -------------------------------------

// GetBasicRedirectStatus Возвращает статус базовой переадресации
// id - Идентификатор, мобильный или добавочный номер абонента
func GetBasicRedirectStatus(id string) int {

}

// TurnOnBasicRedirect Включает базовую переадресацию
// id - Идентификатор, мобильный или добавочный номер абонента
// br - Номера для переадресации
func TurnOnBasicRedirect(id string, br BasicRedirect) error {

}

// TurnOffBasicRedirect Отключает базовую переадресацию
// id - Идентификатор, мобильный или добавочный номер абонента
func TurnOffBasicRedirect(id string) error {

}

//  ------------------------------------- Выборочная переадресация вызовов -------------------------------------

// Возвращает список правил выборочной переадресации
// id - Идентификатор, мобильный или добавочный номер абонента
func GetSelectiveCallRules(id string) (CfsStatusResponse, error) {

}

// Добавляет правило для выборочной переадресации
// id - Идентификатор, мобильный или добавочный номер абонента
// rule -Запрос для добавления правила
func AddSelectiveCallRule(id string, rule CfsRuleUpdate) error {

}

// Включает выборочную переадресацию
// id - Идентификатор, мобильный или добавочный номер абонента
func TurnOnSelectiveRedirect(id string) error {

}

// UpdateSelectiveCallRule Обновляет правило
// id - Идентификатор, мобильный или добавочный номер абонента
// ruleID -Идентификатор правила
// rule - Запрос для обновления правила
func UpdateSelectiveCallRule(id string, ruleID int, rule CfsRuleUpdate) error {

}

// TurnOffSelectiveRedirect Отключает выборочную переадресацию
// id - Идентификатор, мобильный или добавочный номер абонента
func TurnOffSelectiveRedirect(id string) error {

}

// DeleteSelectiveRedirect Удаляет правило
// id - Идентификатор, мобильный или добавочный номер абонента
// ruleID - Идентификатор правила
func DeleteSelectiveCallRule(id string, ruleID int) error {

}

//  ------------------------------------- Выборочный прием звонков -------------------------------------

// Статус и список правил для выборочного приема звонков
// id - Идентификатор, мобильный или добавочный номер абонента
func IncCallRules(id string) (BwlStatusResponse, error) {

}

// AddIncCallRule Добавляет правило для выборочного приема звонков
// id - Идентификатор, мобильный или добавочный номер абонента
// ruleUpdate - Запрос для добавления правила
func AddIncCallRule(id string, rule BwlRuleAdd, ruleUpdate BwlRuleUpdate) int {

}

// TurnOnSelectiveCallReceive Включает выборочный прием звонков
// id - Идентификатор, мобильный или добавочный номер абонента
// t - Тип правила
func TurnOnSelectiveCallReceive(id string, t int) error {
}

// UpdateSelectiveReceiveRule Обновляет правило для выборочного приема звонков
// id - Идентификатор, мобильный или добавочный номер абонента
// ruleID - Идентификатор правила
// ruleUpdate -Запрос для обновления правила
func UpdateSelectiveReceiveRule(id string, ruleID int, ruleUpdate BwlRuleUpdate) error {

}

// TurnOffSelectiveReceiveRule Отключает выборочный прием звонков
// id - Идентификатор, мобильный или добавочный номер абонента
func TurnOffSelectiveReceiveRule(id string) error {

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
func GetRecords(id string) ([]CallRecord, error) {

}

// DeleteRecord Удаляет запись разговора по уникальному идентификатору записи recordId.
// id - Идентификатор записи разговора
func DeleteRecord(id string) error {

}

// GetRecordInfo Возвращает запись разговора по уникальному идентификатору записи recordId.
// id - Идентификатор записи разговора
func GetRecordInfo(id string) (CallRecord, error) {

}

// GetRecordInfoFromEvent Возвращает запись разговора по ID разговора из события и ID пользователя из того же события.
// id - Идентификатор разговора из события
// userId - Идентификатор пользователя из события
func GetRecordInfoFromEvent(id string, userId string) (CallRecord, error) {

}

// GetRecordFile Возвращает файл записи разговора по уникальному идентификатору записи recordId
// id - Идентификатор разговора из события

func GetRecordFile(id string) (Reader, error) {

}

// GetRecordFileFromEvent Возвращает запись разговора по ID разговора  из события и ID пользователя из того же события.
// id - Идентификатор разговора из события
// userId - Идентификатор пользователя из события

func GetRecordFileFromEvent(id string, userId string) (Reader, error) {

}

//  ------------------------------------- Операции со входящими номерами  -------------------------------------

// GetAllIncNumbers Возвращает список всех входящих номеров
// id - Идентификатор разговора из события
// userId - Идентификатор пользователя из события

func GetAllIncNumbers() ([]NumberInfo, error) {

}

// FindIncNumberById Ищет входящий номер по идентификатору, номеру или добавочному номеру
// id - Идентификатор, номер или добавочный номер

func FindIncNumberById(id string) (NumberInfo, error) {

}

//  ------------------------------------- Подписка на Xsi-Events  -------------------------------------

// XSIEventSubscribtion Формирует подписку на Xsi-Events
// Подписка может быть использована для интеграции со сторонними системами, которым необходим контроль над звонками абонентов облачной АТС в реальном времени.
// API использует механизм подписки на события, ассоциированные с тем или иным абонентом, номером или всем клиентом.
// Например, Абонент облачной АТС принимает вызов, сторонняя CRM система получает обновления о текущем статусе вызова (ringing, established, completed).
// req - Запрос для подписки на события

func XSIEventSubscription(reg SubscriptionRequest) (SubscriptionResult, error) {

}

// GetXSIEventSubscriptionInfo Возвращает информацию о подписке на Xsi-Events
// id - Идентификатор подписки

func GetXSIEventSubscriptionInfo(id string) (SubscriptionInfo, error) {

}

// TurnOffXSIEventSubscription Отключает подписку на Xsi-Events
// id - Идентификатор отключаемой подписки

func TurnOffXSIEventSubscription(id string) error {

}
