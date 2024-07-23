package db

import (
	"bflog/config"
	"bflog/utils"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

// DNSLog 数据库表的结构体
type Dnslog struct {
	ID          int       `json:"id"`
	ReceiveIP   string    `json:"receiveip"`
	QueryName   string    `json:"queryname"`
	QueryType   string    `json:"querytype"`
	CreatedTime time.Time `json:"createtime"`
}

type DnsRule struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	IPAddresses string `json:"ip_addresses"`
}

type User struct {
	ID       int    `gorm:"primaryKey"`
	Username string `gorm:"size:255;unique;not null"`
	Password string `gorm:"size:255;not null"`
}

type HttpResponse struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	Path        string `json:"path"`
	StatusCode  string `json:"statuscode"`
	RedirectUrl string `json:"redirecturl"`
	Header      string `json:"header"`
	Body        string `json:"body"`
	Method      string `json:"method"`
	//CreatedAt   time.Time `json:"createtime"`
}

type HttpRequestLog struct {
	ID         uint      `json:"id"`
	Hostname   string    `json:"hostname"`
	Timestamp  time.Time `json:"timestamp"`
	RemoteAddr string    `json:"remoteaddr"`
	Method     string    `json:"method"`
	URL        string    `json:"url"`
	Header     string    `json:"header"`
	Body       string    `json:"body"`
	Path       string    `json:"path"`
}

// DBClient 封装数据库客户端的结构体
type DBClient struct {
	Client   *gorm.DB
	InsertCh chan Dnslog // 通道用于传递要插入的记录
}

// 全局 DBClient 实例
var dbClient *DBClient

// InitDB 初始化数据库
func InitDB() {
	logrus.Info("load db")

	// 配置数据库连接
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.GetBase().Database.Username, config.GetBase().Database.Password,
		config.GetBase().Database.Host, config.GetBase().Database.Port, config.GetBase().Database.DBName)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		AllowGlobalUpdate:      true,
		PrepareStmt:            true,
		NamingStrategy:         schema.NamingStrategy{SingularTable: true},
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		logrus.Fatalf("failed to connect to database: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		logrus.Fatalf("failed to get sqlDB from gorm.DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 创建全局 DBClient 实例
	dbClient = &DBClient{
		Client:   db,
		InsertCh: make(chan Dnslog, 100), // 初始化带缓冲区的通道
	}

	// 启动异步插入
	go dbClient.asyncInsertWorker()
}

// GetDB 返回全局 DBClient 实例
func GetDB() *DBClient {
	return dbClient
}

// AsyncInsertWorker 处理从通道中接收的 DNS 记录并插入到数据库中
func (client *DBClient) asyncInsertWorker() {
	for record := range client.InsertCh {
		//logrus.Infof("Inserting record into database: %+v", record)
		result := client.Client.Create(&record)
		if result.Error != nil {
			logrus.Error(result.Error)
		}
	}
}

// InsertRecord 异步将 DNS 记录发送到通道
func (client *DBClient) InsertRecord(record Dnslog) {
	select {
	case client.InsertCh <- record: // 发送记录到通道
		logrus.Infof("Record inserted into channel: %+v", record)
	default:
		logrus.Warnf("Insert channel is full, dropping record: %+v", record)
	}
}

// 优雅关闭通道，在需要关闭时调用
func (client *DBClient) Close() {
	close(client.InsertCh)
	logrus.Info("Insert channel closed.")
}

func (client *DBClient) InsertLog(log HttpRequestLog) error {
	return client.Client.Create(&log).Error
}

func (client *DBClient) GetHttpResponse(path string, method string) (*HttpResponse, error) {
	var response HttpResponse
	result := client.Client.Where("path = ? and method = ?", path, method).First(&response)
	if result.Error != nil {
		return nil, result.Error
	}
	return &response, nil
}

// 查询dnslog
func (client *DBClient) GetDnslog(receiveIP string, queryName string, queryType string, filter *utils.PaginationAndTimeFilter) ([]Dnslog, int, error) {
	var logs []Dnslog
	var totalCount int64
	query := client.Client.Model(&Dnslog{})

	// 添加过滤条件
	if receiveIP != "" {
		query = query.Where("receive_ip LIKE ?", "%"+receiveIP+"%")
	}
	if queryName != "" {
		query = query.Where("query_name LIKE ?", "%"+queryName+"%")
	}
	if queryType != "" {
		query = query.Where("query_type LIKE ?", "%"+queryType+"%")
	}

	countQuery := query.Session(&gorm.Session{})

	// 获取总条目数
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	query = utils.ApplyPaginationAndTimeFilter(query, filter)

	// 获取总数（用于分页）

	// 查询数据
	if err := query.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, int(totalCount), nil
}

func (client *DBClient) GetUserByUsername(username string) (*User, error) {
	var user User
	result := client.Client.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (client *DBClient) InsertUser(user User) error {
	return client.Client.Create(&user).Error
}

func (client *DBClient) Deletelog(id int) error {
	if client.Client == nil {
		return errors.New("database client is not initialized")
	}
	return client.Client.Delete(&Dnslog{}, "id = ?", id).Error
}

func (client *DBClient) DeleteDnsLogsByIds(ids []int) error {
	// 使用 GORM 的 Delete 方法和 WHERE 子句批量删除
	return client.Client.Where("id IN (?)", ids).Delete(&Dnslog{}).Error
}

func (client *DBClient) DeleteDnsLogsByName(queryname string) error {
	return client.Client.Where("query_name  = ?", queryname).Delete(&Dnslog{}).Error
}

func (client *DBClient) DeleteHttplog(id int) error {
	if client.Client == nil {
		return errors.New("database client is not initialized")
	}
	return client.Client.Delete(&HttpRequestLog{}, "id = ?", id).Error
}

func (client *DBClient) DeleteHttprule(id int) error {
	if client.Client == nil {
		return errors.New("database client is not initialized")
	}
	return client.Client.Delete(&HttpResponse{}, "id = ?", id).Error
}

func (client *DBClient) DeleteDnsrule(id int) error {
	if client.Client == nil {
		return errors.New("database client is not initialized")
	}
	return client.Client.Delete(&DnsRule{}, "id = ?", id).Error
}

func (client *DBClient) DeleteHttpLogsByIds(ids []int) error {
	// 使用 GORM 的 Delete 方法和 WHERE 子句批量删除
	return client.Client.Where("id IN (?)", ids).Delete(&HttpRequestLog{}).Error
}

func (client *DBClient) GetHttplog(hostname string, remoteaddr string, method string, url string, header string, body string, path string, filter *utils.PaginationAndTimeFilter) ([]HttpRequestLog, int, error) {
	var logs []HttpRequestLog
	query := client.Client.Model(&HttpRequestLog{})
	var totalCount int64
	// 添加过滤条件
	if hostname != "" {
		query = query.Where("hostname LIKE ?", "%"+hostname+"%")
	}
	if remoteaddr != "" {
		query = query.Where("remoteaddr LIKE ?", "%"+remoteaddr+"%")
	}
	if method != "" {
		query = query.Where("method LIKE ?", "%"+method+"%")
	}
	if url != "" {
		query = query.Where("url LIKE ?", "%"+url+"%")
	}
	if header != "" {
		query = query.Where("header LIKE ?", "%"+header+"%")
	}
	if body != "" {
		query = query.Where("body LIKE ?", "%"+body+"%")
	}
	if path != "" {
		query = query.Where("path LIKE ?", "%"+path+"%")
	}

	// 添加时间段条件
	countQuery := query.Session(&gorm.Session{})

	// 获取总条目数
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	query = utils.ApplyPaginationAndTimeFilter(query, filter)

	// 查询数据
	if err := query.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, int(totalCount), nil
}

func (client *DBClient) GetHttprule(filter *utils.HttpRuleFilter) ([]HttpResponse, int, error) {
	var rules []HttpResponse
	query := client.Client.Model(&HttpResponse{})
	var totalCount int64
	if filter.Hostname != "" {
		query = query.Where("hostname LIKE ?", "%"+filter.Hostname+"%")
	}
	if filter.Remoteaddr != "" {
		query = query.Where("remoteaddr LIKE ?", "%"+filter.Remoteaddr+"%")
	}
	if filter.Method != "" {
		query = query.Where("method LIKE ?", "%"+filter.Method+"%")
	}
	if filter.Url != "" {
		query = query.Where("url LIKE ?", "%"+filter.Url+"%")
	}
	if filter.Header != "" {
		query = query.Where("header LIKE ?", "%"+filter.Header+"%")
	}
	if filter.Body != "" {
		query = query.Where("body LIKE ?", "%"+filter.Body+"%")
	}
	if filter.Path != "" {
		query = query.Where("path LIKE ?", "%"+filter.Path+"%")
	}
	countQuery := query.Session(&gorm.Session{})
	// 获取总条目数
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}
	paginationFilter := &utils.PaginationAndTimeFilter{
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		StartTime: filter.StartTime,
		EndTime:   filter.EndTime,
	}
	query = utils.ApplyPaginationAndTimeFilter(query, paginationFilter)

	// 查询数据
	if err := query.Find(&rules).Error; err != nil {
		return nil, 0, err
	}

	return rules, int(totalCount), nil
}

func (client *DBClient) Getdnsrule(name string, filter *utils.PaginationAndTimeFilter) ([]DnsRule, int, error) {
	var result []DnsRule
	query := client.Client.Model(&DnsRule{})
	var totalCount int64
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}
	paginationFilter := &utils.PaginationAndTimeFilter{
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		StartTime: filter.StartTime,
		EndTime:   filter.EndTime,
	}
	query = utils.ApplyPaginationAndTimeFilter(query, paginationFilter)

	// 查询数据
	if err := query.Find(&result).Error; err != nil {
		return nil, 0, err
	}

	return result, int(totalCount), nil
}

func (client *DBClient) adddnsrule(rule DnsRule) error {
	return client.Client.Create(&rule).Error
}

func (client *DBClient) deletednsrule(id int) error {
	return client.Client.Delete(&DnsRule{}, "id = ?", id).Error
}

func (client *DBClient) DeleteAllHttpLogs() interface{} {
	return client.Client.Unscoped().Delete(&HttpRequestLog{}).Error
}

func (client *DBClient) DeleteAllDnsLogs() interface{} {
	return client.Client.Unscoped().Delete(&Dnslog{}).Error
}
