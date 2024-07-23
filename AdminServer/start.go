package AdminServer

import (
	"bflog/config"
	"encoding/json"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"net/http"
)

type JsonResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ApiResponse struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data,omitempty"`
}

// DataResponse 定义返回数据的结构
type DataResponse struct {
	Items []interface{} `json:"items"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
}

type PaginatedResponse struct {
	CurrentPage int         `json:"currentPage"`
	PageSize    int         `json:"perPage"`
	TotalPages  int         `json:"pager"`
	TotalCount  int         `json:"total"`
	Data        interface{} `json:"data"`
}

func handleAuth(w http.ResponseWriter, r *http.Request) bool {
	if config.GetBase().Dev == 1 {
		return true
	}
	authHeader := r.Header.Get("Auth-Token")

	if authHeader == "" {
		sendJSONResponse(w, 1, "Unauthorized", nil)
		return false
	}
	_, err := ValidateJWT(authHeader)

	if err != nil {
		sendJSONResponse(w, 1, "Unauthorized", nil)
		return false
	}
	return true
}

func sendJSONResponse(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := JsonResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response to JSON", http.StatusInternalServerError)
	}
}

func setupRoutes() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/dnslogs", getDnslogs)
	mux.HandleFunc("/api/login", loginHandler)
	mux.HandleFunc("/api/deldnslogbyid", deleteDnslog)
	mux.HandleFunc("/api/deldnslogbyids", deleteDnsLogsByIds)
	mux.HandleFunc("/api/deldnslogbyname", deleteDnsLogsByname)
	mux.HandleFunc("/api/httplogs", getHttplogs)
	mux.HandleFunc("/api/delhttplogbyid", deleteHttplog)
	mux.HandleFunc("/api/delhttplogbyids", deleteHttpLogsByIds)
	mux.HandleFunc("/api/gethttprule", getHttprules)
	mux.HandleFunc("/api/delhttprule", deleteHttprule)
	mux.HandleFunc("/api/updatehttprule", updateHttprule)
	mux.HandleFunc("/api/addhttprule", AddHttpResponse)
	mux.HandleFunc("/api/getdnsrule", getdnsrule)
	mux.HandleFunc("/api/adddnsrule", adddnsrule)
	mux.HandleFunc("/api/updatednsrule", updateDnsRule)
	mux.HandleFunc("/api/deldnsrulebyid", deleteDnsRule)
	mux.HandleFunc("/api/delallhttp", deleteAllHttpLogs)
	mux.HandleFunc("/api/delalldns", deleteAllDnsLogs)
	port := ":" + config.GetBase().Server.Adminport

	// 设置 CORS 中间件
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://127.0.0.1:8888", "http://localhost:4000", config.GetBase().Server.Admindomain}, // 允许的来源
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},                                             // 允许的 HTTP 方法
		AllowedHeaders:   []string{"*"},                                                                                   // 允许的头
		AllowCredentials: true,                                                                                            // 允许凭证
	})

	// 包装 HTTP 处理器
	handler := c.Handler(mux)
	if err := http.ListenAndServe(port, handler); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}

}

func Start() error {
	setupRoutes()

	return nil
}
