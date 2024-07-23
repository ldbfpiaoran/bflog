package DnsServer

import (
	"bflog/config"
	"bflog/db"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"strings"
	"time"
)

// removeTrailingDot 移除域名末尾的点
func removeTrailingDot(name string) string {
	return strings.TrimSuffix(name, ".")
}

func GetNextIPAddress(domain string) string {

	ip, _ := db.GetRedis().LPop(domain)
	if ip == "" {
		return config.GetBase().Server.Defaultip
	}
	if ip != "" {
		db.GetRedis().RPush(domain, ip)
	}
	return ip
}

func InsertRecord(record db.Dnslog) {
	select {
	case db.GetDB().InsertCh <- record:

	}
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	// 创建响应消息
	msg := dns.Msg{}
	msg.SetReply(r)

	receiveIP := w.RemoteAddr().String()
	ttl := 60

	//logrus.Info(receiveIP)
	// 记录请求
	for _, q := range r.Question {
		if strings.HasSuffix(q.Name, config.GetBase().Server.Subdomain) {
			//logrus.Info(config.GetBase().Server.Subdomain)
			defaultip := config.GetBase().Server.Defaultip
			//logrus.Info(defaultip)
			domain := removeTrailingDot(q.Name)
			ip := GetNextIPAddress(domain)
			if ip != "" {
				defaultip = ip
				ttl = 0
			}
			record := db.Dnslog{
				ReceiveIP:   receiveIP,
				QueryName:   domain,
				QueryType:   dns.TypeToString[q.Qtype],
				CreatedTime: time.Now(),
			}
			InsertRecord(record)
			switch q.Qtype {
			case dns.TypeA:
				rr := &dns.A{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: q.Qtype,
						Class:  dns.ClassINET,
						Ttl:    uint32(ttl),
					},
					A: net.ParseIP(defaultip),
				}
				msg.Answer = append(msg.Answer, rr)
			}
		}
	}

	// 发送响应
	w.WriteMsg(&msg)
}

func Start() error {
	logrus.Info("start dns server ")

	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		go handleDNSRequest(w, r) // 使用 goroutine 处理每个请求
	})
	server := &dns.Server{Addr: ":53", Net: "udp"}
	log.Printf("启动 DNS 服务器，监听 %s\n", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("无法启动 DNS 服务器: %v\n", err)
	}
	return nil
}
