package main

import (
	"bflog/config"
	"bflog/db"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 初始化数据库连接
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}
	// 创建一个用户
	db.InitDB()
	username := "admin"

	password := "123456" // 更复杂的密码

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Failed to generate hashed password: %v\n", err)
		return
	}
	logrus.Info(string(hashedPassword))

	// 构造用户对象
	user := db.User{
		Username: username,
		Password: string(hashedPassword),
	}
	//
	//// 插入用户到数据库
	err = db.GetDB().InsertUser(user)
	if err != nil {
		fmt.Printf("Failed to insert user into database: %v\n", err)
		return
	}
	//
	fmt.Println("User inserted successfully:", username)
}
