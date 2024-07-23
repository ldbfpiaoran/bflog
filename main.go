package main

import (
	"bflog/AdminServer"
	"bflog/DnsServer"
	"bflog/HttpServer"
	"bflog/config"
	"bflog/db"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func init() {

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		TimestampFormat:           "2006-01-02 15:04:05", //时间格式
		FullTimestamp:             true,
	})
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true)

}

func start() {
	ctx, cancel := context.WithCancel(context.Background())
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}
	db.InitRedisDB()
	db.InitDB()

	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		db.GetDB().Close()
		fmt.Println("server stop")
		cancel()
	}()
	go AdminServer.Start()
	go DnsServer.Start()
	go HttpServer.Start()
	<-ctx.Done()

}

func initOnly(username, password string) {
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}
	db.InitDB()
	fmt.Printf("初始化时使用的用户名: %s, 密码: %s\n", username, password)
	// 这里可以添加更多的初始化逻辑，比如保存这些信息到数据库
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Failed to generate hashed password: %v\n", err)
		return
	}
	user := db.User{
		Username: username,
		Password: string(hashedPassword),
	}
	err = db.GetDB().InsertUser(user)
	if err != nil {
		fmt.Printf("Failed to insert user into database: %v\n", err)
		return
	}
	fmt.Println("User inserted successfully:", username)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("请提供一个参数: start 或 init")
	}

	arg := os.Args[1]
	switch arg {
	case "start":
		start()
	case "init":
		if len(os.Args) < 4 {
			log.Fatal("请提供 username 和 password 参数")
		}
		username := os.Args[2]
		password := os.Args[3]
		// 执行 init 参数的逻辑

		initOnly(username, password)
	}

}
