package AdminServer

import (
	"bflog/db"
	"bflog/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

func getHttprules(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	}
	var filter utils.HttpRuleFilter
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	responses, totalCount, err := db.GetDB().GetHttprule(&filter)
	if err != nil {
		sendJSONResponse(w, 1, err.Error(), nil)
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	interfaceLogs := utils.ConvertToInterfaceSlice(responses)
	data := DataResponse{
		Items: interfaceLogs,
		Total: totalCount,
		Page:  filter.Page,
	}
	sendJSONResponse(w, 0, "success", data)
}

func deleteHttprule(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	} // 认证请求

	// 获取查询参数 ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		sendJSONResponse(w, 1, "参数 缺失", nil)
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
	err = db.GetDB().DeleteHttprule(id)
	if err != nil {
		sendJSONResponse(w, 1, "删除失败", nil)
		//http.Error(w, "Failed to delete DNS log", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, 0, "DNS log deleted successfully", nil)
}

func updateHttprule(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	} // 认证请求

	var httpResponse db.HttpResponse

	// 从请求体中解码 JSON 数据
	if err := json.NewDecoder(r.Body).Decode(&httpResponse); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if httpResponse.ID == 0 {
		http.Error(w, "ID is required for updating", http.StatusBadRequest)
		return
	}
	updateData := map[string]interface{}{
		"Method":      httpResponse.Method,
		"Path":        httpResponse.Path,
		"StatusCode":  httpResponse.StatusCode,
		"Body":        httpResponse.Body,
		"Header":      httpResponse.Header,
		"RedirectUrl": httpResponse.RedirectUrl,
	}
	if err := db.GetDB().Client.Model(&db.HttpResponse{}).Where("id = ?", httpResponse.ID).Updates(updateData).Error; err != nil {
		http.Error(w, "Failed to update HTTP response", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, 0, "update success", nil)
}

type HttpResponsePayload struct {
	Path        string `json:"path,omitempty"`
	RedirectUrl string `json:"redirecturl,omitempty"`
	Header      string `json:"header,omitempty"`
	Body        string `json:"body,omitempty"`
	Method      string `json:"method,omitempty"`
	StatusCode  string `json:"statuscode"`
}

func AddHttpResponse(w http.ResponseWriter, r *http.Request) {
	if !handleAuth(w, r) { // 添加 CORS 头部，并检查授权
		return
	} // 认证请求

	var payload HttpResponsePayload

	// 从请求体中解码 JSON 数据
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var existingResponse db.HttpResponse
	if err := db.GetDB().Client.Where("path = ? and method =?", payload.Path, payload.Method).First(&existingResponse).Error; err == nil {
		sendJSONResponse(w, 1, "path已存在", nil)
		return
	}

	// 设置创建和更新时间
	httpResponse := db.HttpResponse{
		Path:        payload.Path,
		RedirectUrl: payload.RedirectUrl,
		StatusCode:  payload.StatusCode,
		Header:      payload.Header,
		Body:        payload.Body,
		Method:      payload.Method,
	}

	// 插入数据库
	if err := db.GetDB().Client.Create(&httpResponse).Error; err != nil {
		http.Error(w, "Failed to add HTTP response", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, 0, "添加成功", nil)
}
