package HttpServer

import (
	"bflog/config"
	"bflog/db"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func formatHeadersToJSON(headers map[string][]string) (string, error) {
	jsonBytes, err := json.Marshal(headers)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func parseJSONToHeaders(jsonStr string) (map[string][]string, error) {
	// 用于存储临时解析结果的通用map
	var tempMap map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &tempMap)
	if err != nil {
		return nil, err
	}

	headers := make(map[string][]string)

	// 遍历临时map，处理不同类型的值
	for key, value := range tempMap {
		switch v := value.(type) {
		case string:
			// 如果是字符串，转换为单元素的字符串数组
			headers[key] = []string{v}
		case []interface{}:
			// 如果是数组，尝试转换为字符串数组
			strSlice := make([]string, len(v))
			for i, item := range v {
				str, ok := item.(string)
				if !ok {
					return nil, errors.New("invalid array element type, expected string")
				}
				strSlice[i] = str
			}
			headers[key] = strSlice
		default:
			return nil, errors.New("invalid value type, expected string or []string")
		}
	}

	return headers, nil
}

func isAllowedDomain(host string, allowedDomains []string) bool {
	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		// 如果没有端口信息，直接使用 host
		hostname = host
	}
	for _, domain := range allowedDomains {
		// 如果是通配符域名
		if strings.HasPrefix(domain, ".") {
			// 去掉通配符部分，只检查主域名是否匹配
			if strings.HasSuffix(hostname, domain[1:]) {
				return true
			}
		} else {
			// 完全匹配检查
			if hostname == domain {
				return true
			}
		}
	}
	return false
}

func httpStatusCode(code string) (int, error) {
	statusCode, err := strconv.Atoi(code)
	if err != nil || statusCode < 100 || statusCode > 599 {
		return 0, fmt.Errorf("invalid status code")
	}
	return statusCode, nil
}

func logRequestHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	hostname := r.Host
	allowedDomains := strings.Split(config.GetBase().Server.ListenDomain, ",")
	if !isAllowedDomain(hostname, allowedDomains) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	method := r.Method
	url := r.URL.String()
	path := r.URL.Path
	remoteAddr := r.RemoteAddr
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read request body", http.StatusInternalServerError)
		return
	}
	bodyString := string(bodyBytes)
	headerJSON, _ := formatHeadersToJSON(r.Header)
	if config.GetBase().Nginx == 1 {
		remoteAddr = strings.Split(r.Header.Get("X-Real-Ip"), ",")[0]
	}
	if remoteAddr == "-" {
		http.Error(w, "Cannot read request body", http.StatusInternalServerError)
		return
	}
	if remoteAddr == "" {
		http.Error(w, "Cannot read request body", http.StatusInternalServerError)
		return
	}
	httpRequestLog := db.HttpRequestLog{
		Hostname:   hostname,
		Timestamp:  time.Now(),
		RemoteAddr: remoteAddr,
		Method:     method,
		URL:        url,
		Header:     headerJSON,
		Body:       bodyString,
		Path:       path,
	}
	if err := db.GetDB().InsertLog(httpRequestLog); err != nil {
		logrus.Errorf("Failed to insert log into database: %v", err)
	}
	decodedPath, err := base64.URLEncoding.DecodeString(path[1:]) // 去掉前导的 '/'
	if err == nil {
		decodedPathStr := string(decodedPath)

		// 检查解码后的内容是否符合 302 重定向的格式
		parts := strings.SplitN(decodedPathStr, ":", 2)
		if len(parts) == 2 {
			// 检查状态码部分是否是有效的 HTTP 状态码
			statusCode := parts[0]
			redirectURL := parts[1]
			if statusCode == "302" {
				http.Redirect(w, r, redirectURL, http.StatusFound)
				return
			} else if statusCode == "301" {
				http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
				return
			} else if code, err := httpStatusCode(statusCode); err == nil {
				// 如果是其他有效的状态码，返回该状态码
				w.WriteHeader(code)
				w.Write([]byte(redirectURL))
				return
			} else {
				logrus.Warnf("Invalid status code: %s", statusCode)
			}
		}
	}
	//  todo 通配符path  参数解析
	responseConfig, _ := db.GetDB().GetHttpResponse(path, method)
	if responseConfig != nil {
		// 设置响应头
		responseHeaders, _ := parseJSONToHeaders(responseConfig.Header)
		if responseHeaders != nil {
			for key, values := range responseHeaders {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}
		}
		//w.Header().Set("Content-Type", "application/json")

		// 如果存在重定向 URL
		if responseConfig.RedirectUrl != "" {
			http.Redirect(w, r, responseConfig.RedirectUrl, http.StatusFound)
			return
		}

		// 设置状态码
		//logrus.Info(responseConfig)
		statusCode, err := strconv.Atoi(responseConfig.StatusCode)
		if err != nil {
			http.Error(w, "StatusCode must be a valid integer", http.StatusBadRequest)
			return
		}
		w.WriteHeader(statusCode)

		// 返回响应数据
		_, _ = w.Write([]byte(responseConfig.Body))
		return
	} else {
		// 默认响应
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Request logged\n"))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Request logged\n"))
	// todo 默认返回值
	return
}

func Start() error {
	http.HandleFunc("/", logRequestHandler)
	port := ":" + config.GetBase().Server.Port

	err := http.ListenAndServe(port, nil)
	if err != nil {
		logrus.Fatalf("Error starting server: %v\n", err)
	}
	return nil
}
