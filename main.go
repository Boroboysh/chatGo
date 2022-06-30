package main

import (
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ConnectUser struct {
	Websocket *websocket.Conn
	ClientIP  string
}

func newConnectUser(ws *websocket.Conn, clientIP string) *ConnectUser {
	return &ConnectUser{
		Websocket: ws,
		ClientIP:  clientIP,
	}
}

func HtmlHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/index.html")
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var users = make(map[ConnectUser]int)

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		if err := ws.Close(); err != nil {
			log.Println("Ошибка: подключение веб-сокета не закрыто", err.Error())
		}
	}()

	log.Println("Клиент подключен:", ws.RemoteAddr().String())
	var socketClient *ConnectUser = newConnectUser(ws, ws.RemoteAddr().String())
	users[*socketClient] = 0
	log.Println("Количество подключенных пользователей ...", len(users))

	for {
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("Ожидание отключения...", err.Error())
			delete(users, *socketClient)
			log.Println("Количество подключенных пользователей ...", len(users))
			return
		}

		for client := range users {
			if err = client.Websocket.WriteMessage(messageType, message); err != nil {
				log.Println("Ошибка отправки сообщения ", client.ClientIP, err.Error())
			}
		}

	}
}

func main() {
	http.HandleFunc("/", HtmlHandler)
	http.HandleFunc("/ws", WebsocketHandler)

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
