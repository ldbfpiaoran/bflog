package AdminServer

import (
	"bflog/db"
	"bflog/utils"
	"encoding/json"
	"net/http"
	"strings"
)

type dnsrule struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Ipaddresss string `json:"ip_addresses"`
}

func getdnsrule(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	}
	name := r.URL.Query().Get("name")
	filter, err := utils.ParsePaginationAndTimeFilter(r)
	if err != nil {
		sendJSONResponse(w, 1, "Invalid pagination or time filter parameters.", nil)
		return
	}
	logs, totalCount, err := db.GetDB().Getdnsrule(name, filter)
	if err != nil {
		sendJSONResponse(w, 1, err.Error(), nil)
		return
	}
	interfaceLogs := utils.ConvertToInterfaceSlice(logs)
	data := DataResponse{
		Items: interfaceLogs,
		Total: totalCount,
		Page:  filter.Page,
	}
	// 设置响应头为 JSON 并返回数据
	sendJSONResponse(w, 0, "success", data)
}

func adddnsrule(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	}
	var dns dnsrule
	if err := json.NewDecoder(r.Body).Decode(&dns); err != nil {
		sendJSONResponse(w, 1, "Invalid request payload", nil)
		return
	}

	var existingRule db.DnsRule
	if err := db.GetDB().Client.Where("name =?", dns.Name).First(&existingRule).Error; err == nil {
		sendJSONResponse(w, 1, "name已存在", nil)
		return
	}
	dnsrule := db.DnsRule{
		Name:        dns.Name,
		IPAddresses: dns.Ipaddresss,
	}
	if err := db.GetDB().Client.Create(&dnsrule).Error; err != nil {
		sendJSONResponse(w, 1, "添加错误", nil)
		return
	}
	data := strings.Split(dns.Ipaddresss, ",")
	const maxRetries = 3
	var err error
	for i := 0; i < maxRetries; i++ {
		err = db.GetRedis().RPush(dns.Name, data)
		if err == nil {
			break // 成功写入 Redis
		}
	}
	if err != nil {
		sendJSONResponse(w, 1, "添加到 Redis 失败，已重试多次", nil)
		return
	}
	sendJSONResponse(w, 0, "添加成功", nil)
}

func updateDnsRule(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	}
	var dns dnsrule
	if err := json.NewDecoder(r.Body).Decode(&dns); err != nil {
		sendJSONResponse(w, 1, "json解析失败", nil)
		return
	}

	var updatedDnsRule db.DnsRule
	tx := db.GetDB().Client.Begin()
	var existingRule db.DnsRule
	if err := tx.Where("id = ?", dns.Id).First(&updatedDnsRule).Error; err != nil {
		tx.Rollback()
		http.Error(w, "DNS rule not found", http.StatusNotFound)
		return
	}
	existingRule.IPAddresses = dns.Ipaddresss
	existingRule.ID = updatedDnsRule.ID
	existingRule.Name = dns.Name
	if err := tx.Save(&existingRule).Error; err != nil {
		tx.Rollback()
		sendJSONResponse(w, 1, "更新失败 ", nil)
		//http.Error(w, "Failed to update DNS rule in MySQL", http.StatusInternalServerError)
		return
	}
	data := strings.Split(dns.Ipaddresss, ",")
	if err := db.GetRedis().Del(dns.Name); err != nil {
		tx.Rollback()
		sendJSONResponse(w, 1, "更新到redis失败 ", nil)
		return
	}
	if err := db.GetRedis().RPush(dns.Name, data); err != nil {
		tx.Rollback()
		sendJSONResponse(w, 1, "更新到redis失败 ", nil)
		return
	}
	if err := tx.Commit().Error; err != nil {
		sendJSONResponse(w, 1, "Failed to commit transaction ", nil)
		return
	}
	sendJSONResponse(w, 0, "更新成功", nil)
}

func deleteDnsRule(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	}

	name := r.URL.Query().Get("id")
	id := r.URL.Query().Get("id")
	// 启动事务

	tx := db.GetDB().Client.Begin()

	// 删除 MySQL 中的数据
	if err := tx.Where("id = ?", id).Delete(&db.DnsRule{}).Error; err != nil {
		tx.Rollback()
		sendJSONResponse(w, 1, "删除失败", nil)
		//http.Error(w, "Failed to delete DNS rule from MySQL", http.StatusInternalServerError)
		return
	}

	// 删除 Redis 中的数据
	if err := db.GetRedis().Del(name); err != nil {
		tx.Rollback()
		sendJSONResponse(w, 1, "redis删除失败", nil)
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		sendJSONResponse(w, 1, "mysql删除失败", nil)
		return
	}

	sendJSONResponse(w, 0, "删除成功", nil)
}
