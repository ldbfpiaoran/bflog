package AdminServer

import (
	"bflog/db"
	"bflog/utils"
	"net/http"
	"strconv"
	"strings"
)

func getHttplogs(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	}

	hostname := r.URL.Query().Get("hostname")
	method := r.URL.Query().Get("method")
	remoteaddr := r.URL.Query().Get("remoteaddr")

	url := r.URL.Query().Get("url")
	header := r.URL.Query().Get("header")
	body := r.URL.Query().Get("body")
	path := r.URL.Query().Get("path")

	filter, err := utils.ParsePaginationAndTimeFilter(r)
	if err != nil {
		sendJSONResponse(w, 1, "分页错误", nil)
		//http.Error(w, "Invalid pagination or time filter parameters.", http.StatusBadRequest)
		return
	}

	logs, totalCount, err := db.GetDB().GetHttplog(hostname, remoteaddr, method, url, header, body, path, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 计算总页数
	interfaceLogs := utils.ConvertToInterfaceSlice(logs)
	data := DataResponse{
		Items: interfaceLogs,
		Total: totalCount,
		Page:  filter.Page,
	}
	// 设置响应头为 JSON 并返回数据
	sendJSONResponse(w, 0, "success", data)

}

func deleteHttplog(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	} // 认证请求

	// 获取查询参数 ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		sendJSONResponse(w, 1, "缺少id", nil)
		//http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSONResponse(w, 1, "错误的id", nil)
		//http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	// 尝试删除记录
	err = db.GetDB().DeleteHttplog(id)
	if err != nil {
		sendJSONResponse(w, 1, "从mysql中删除失败", nil)
		return
	}

	sendJSONResponse(w, 0, "DNS log deleted successfully", nil)
}

func deleteHttpLogsByIds(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	}

	// 获取并解析 ids 参数
	idsParam := r.URL.Query().Get("ids")
	if idsParam == "" {
		sendJSONResponse(w, 1, "ids 参数缺失", nil)
		//http.Error(w, "Missing 'ids' parameter", http.StatusBadRequest)
		return
	}

	// 解析 ids 参数
	idStrings := strings.Split(idsParam, ",")
	ids := make([]int, 0, len(idStrings))
	for _, idStr := range idStrings {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			sendJSONResponse(w, 1, "错误的ids", nil)
			//http.Error(w, "Invalid 'ids' parameter", http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}

	// 调用数据库方法批量删除 DNS 日志
	if err := db.GetDB().DeleteHttpLogsByIds(ids); err != nil {
		sendJSONResponse(w, 1, "删除logs失败", nil)
		//http.Error(w, "Failed to delete DNS logs", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, 0, "success", nil)
}
