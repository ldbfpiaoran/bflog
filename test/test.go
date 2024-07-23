package main

import (
	"bflog/config"
	"bflog/db"
	"fmt"
	"log"
	"net"
	"strings"
)

func isAllowedDomain(host string, allowedDomains []string) bool {
	host, _, err := net.SplitHostPort(host)
	if err != nil {
		// 如果没有端口信息，直接使用 host

	}
	fmt.Println(host)
	for _, domain := range allowedDomains {
		// 如果是通配符域名
		if strings.HasPrefix(domain, ".") {
			// 去掉通配符部分，只检查主域名是否匹配
			if strings.HasSuffix(host, domain[1:]) {
				return true
			}
		} else {
			// 完全匹配检查
			if host == domain {
				return true
			}
		}
	}
	return false
}

func main() {
	// 初始化数据库连接
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}
	// 创建一个用户
	db.InitDB()
	fmt.Println(config.GetBase().Server.ListenDomain)
	allowedDomains := strings.Split(config.GetBase().Server.ListenDomain, ",")
	fmt.Println(!isAllowedDomain("360mrpc.0756nanke.com", allowedDomains))
}
