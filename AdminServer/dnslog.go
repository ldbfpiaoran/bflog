package AdminServer

import (
	"bflog/db"
	"bflog/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func getDnslogs(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 鉴权
		return
	}

	// 获取查询参数
	receiveIP := r.URL.Query().Get("receiveip")
	queryName := r.URL.Query().Get("queryname")
	queryType := r.URL.Query().Get("querytype")

	// 默认分页参数
	filter, err := utils.ParsePaginationAndTimeFilter(r)
	if err != nil {
		sendJSONResponse(w, 1, "Invalid pagination or time filter parameters.", nil)
		return
	}

	// 调用 GetDnslog 方法获取数据
	logs, totalCount, err := db.GetDB().GetDnslog(receiveIP, queryName, queryType, filter)
	if err != nil {
		sendJSONResponse(w, 1, err.Error(), nil)
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	interfaceLogs := utils.ConvertToInterfaceSlice(logs)
	// 创建分页响应
	data := DataResponse{
		Items: interfaceLogs,
		Total: totalCount,
		Page:  filter.Page,
	}
	// 设置响应头为 JSON 并返回数据
	sendJSONResponse(w, 0, "success", data)
}

func deleteDnslog(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	} // 认证请求

	// 获取查询参数 ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		sendJSONResponse(w, 1, "ID parameter is required", nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSONResponse(w, 1, "Invalid ID parameter", nil)
		return
	}

	// 尝试删除记录
	err = db.GetDB().Deletelog(id)
	if err != nil {
		sendJSONResponse(w, 1, "Failed to delete DNS log", nil)
		return
	}

	sendJSONResponse(w, 0, "DNS log deleted successfully", nil)
}

func deleteDnsLogsByIds(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	}

	// 获取并解析 ids 参数
	idsParam := r.URL.Query().Get("ids")
	if idsParam == "" {
		sendJSONResponse(w, 1, "Missing 'ids' parameter", nil)
		return
	}

	// 解析 ids 参数
	idStrings := strings.Split(idsParam, ",")
	ids := make([]int, 0, len(idStrings))
	for _, idStr := range idStrings {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			sendJSONResponse(w, 1, "Invalid 'ids' parameter", nil)
			//http.Error(w, "Invalid 'ids' parameter", http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}

	// 调用数据库方法批量删除 DNS 日志
	if err := db.GetDB().DeleteDnsLogsByIds(ids); err != nil {
		sendJSONResponse(w, 1, "Failed to delete DNS logs", nil)
		//http.Error(w, "Invalid 'ids' parameter", http.StatusBadRequest)
		//http.Error(w, "Failed to delete DNS logs", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, 0, "success", nil)
}

type Deldnslog struct {
	Name string `json:"queryname"`
}

func deleteDnsLogsByname(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) {
		return
	}

	var deldnslog Deldnslog
	if err := json.NewDecoder(r.Body).Decode(&deldnslog); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 调用数据库方法批量删除 DNS 日志
	if err := db.GetDB().DeleteDnsLogsByName(deldnslog.Name); err != nil {
		sendJSONResponse(w, 1, "Failed to delete DNS logs", nil)
		return
	}

	sendJSONResponse(w, 0, "success", nil)
}
