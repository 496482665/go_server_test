package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	// 指示当前用户属于哪个Server
	Server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		Server: server,
	}

	// 监听用户信息
	go user.ListenMessage()

	return user
}

// 监听当前User Channel方法，有消息则发送给对端客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		fmt.Println(msg)
		_, _ = u.conn.Write([]byte(msg + "\n"))

	}
}

// 用户上线功能
func (u *User) Online() {
	// 用户上线
	u.Server.mapLock.Lock()
	u.Server.OnlineMap[u.Name] = u
	u.Server.mapLock.Unlock()

	// 广播用户上线
	u.Server.BroadCast(u, "已上线")
}

// 用户下线功能
func (u *User) Offline() {
	// 用户下线
	u.Server.mapLock.Lock()
	delete(u.Server.OnlineMap, u.Name)
	u.Server.mapLock.Unlock()

	// 广播用户下线
	u.Server.BroadCast(u, "下线")
}

// 用户发送消息功能
func (u *User) SendMessage(msg string) {
	if msg == "who" {
		// 查询当前用户有哪些，返回给特定用户
		for _, user := range u.Server.OnlineMap {
			onlineMsg := "user:" + user.Name
			u.C <- onlineMsg
		}
	} else if len(msg) >= 7 && msg[:7] == "rename:" {
		newName := strings.Split(msg, ":")[1]
		_, ok := u.Server.OnlineMap[newName]
		if ok {
			u.C <- string("用户名已被使用")
		}
		delete(u.Server.OnlineMap, u.Name)
		u.Server.OnlineMap[newName] = u
	} else {
		u.Server.BroadCast(u, msg)
	}
}
