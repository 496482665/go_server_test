package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// NewServer 创建一个Server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		mapLock:   sync.RWMutex{},
		Message:   make(chan string),
	}
	return server
}

func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := user.Name + ":" + msg + "\n"

	s.Message <- sendMsg
}

// 监听message广播消息,并发送给每个用户
func (s *Server) SendToUser() {
	for {
		msg := <-s.Message
		// 读取Message中的消息发给所有用户
		s.mapLock.Lock()
		for _, user := range s.OnlineMap {
			user.C <- msg
		}
		s.mapLock.Unlock()
	}
}

func (s *Server) Handler(conn net.Conn) {
	// 控制创建好的连接
	fmt.Println("链接建立成功")

	user := NewUser(conn, s)

	// 用户上线
	user.Online()

	// 用户发送消息就广播出来
	go func() {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if n == 0 {
			user.Offline()
			return
		}

		if err != nil && err != io.EOF {
			fmt.Println("Conn Read err:", err)
			return
		}

		// 提取用户的消息
		msg := string(buf[:n-1])

		// 将得到的消息广播
		user.SendMessage(msg)
	}()

	// 当前handler阻塞
	select {}
}

// Start 启动服务器
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Printf("构造listener出错了:%s", err)
	}

	// 启动监听广播用户上线
	go s.SendToUser()

	// close listen socket
	defer listener.Close()

	// accept
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("创建监听连接出错了:%s", err)
			continue
		}

		go s.Handler(conn)

	}

}
