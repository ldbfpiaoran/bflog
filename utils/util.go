package utils

import (
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

// PaginationAndTimeFilter 用于承载分页和时间过滤参数
type PaginationAndTimeFilter struct {
	Page      int       `json:"page"`
	PageSize  int       `json:"pageSize"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type HttpRuleFilter struct {
	StatusCode int       `json:"status_code"`
	Body       string    `json:"body"`
	Hostname   string    `json:"hostname"`
	Remoteaddr string    `json:"remoteaddr"`
	Method     string    `json:"method"`
	Url        string    `json:"url"`
	Header     string    `json:"header"`
	Path       string    `json:"path"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Page       int       `json:"page"`
	PageSize   int       `json:"perPage"`
}

// ParsePaginationAndTimeFilter 从请求中解析分页和时间过滤参数
func ParsePaginationAndTimeFilter(r *http.Request) (*PaginationAndTimeFilter, error) {
	page := 1
	pageSize := 10
	var startTime, endTime time.Time
	var err error

	// 解析分页参数
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err == nil && p > 0 {
			page = p
		}
	}

	pageSizeStr := r.URL.Query().Get("perPage")
	if pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// 解析时间段参数
	startTimeStr := r.URL.Query().Get("startTime")
	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, err
		}
	}

	endTimeStr := r.URL.Query().Get("endTime")
	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, err
		}
	}

	return &PaginationAndTimeFilter{
		Page:      page,
		PageSize:  pageSize,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

func ConvertToInterfaceSlice[T any](slice []T) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}

// ApplyPaginationAndTimeFilter 将分页和时间过滤应用到查询中
func ApplyPaginationAndTimeFilter(query *gorm.DB, filter *PaginationAndTimeFilter) *gorm.DB {
	if !filter.StartTime.IsZero() {
		query = query.Where("created_at >= ?", filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	offset := (filter.Page - 1) * filter.PageSize
	return query.Offset(offset).Limit(filter.PageSize)
}
