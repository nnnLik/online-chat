package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Pool struct {
	connections map[*websocket.Conn]bool
	broadcast   chan []byte
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
}

func NewPool() *Pool {
	return &Pool{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *websocket.Conn),
		unregister:  make(chan *websocket.Conn),
	}
}

func (p *Pool) Start() {
    for {
        select {
        case conn := <-p.register:
            p.connections[conn] = true
        case conn := <-p.unregister:
            if _, ok := p.connections[conn]; ok {
                delete(p.connections, conn)
                conn.Close()
            }
        case message := <-p.broadcast:
            for conn := range p.connections {
                err := conn.WriteMessage(websocket.TextMessage, message)
                if err != nil {
                    log.Printf("error sending message to client: %v", err)
                    delete(p.connections, conn)
                    conn.Close()
                }
            }
        }
    }
}

type Client struct {
	pool *Pool
	conn *websocket.Conn
	send chan []byte
}

type message struct {
    id       string
    content  []byte
}

func (c *Client) Read() {
	defer func() {
		c.pool.unregister <- c.conn
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		c.pool.broadcast <- message
	}
}

func (c *Client) Write() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func broadcast(pool map[*websocket.Conn]bool, message []byte) {
    for conn := range pool {
        err := conn.WriteMessage(websocket.TextMessage, message)
        if err != nil {
            log.Printf("error sending message to client: %v", err)
            conn.Close()
            delete(pool, conn)
        }
    }
}

func serveWs(pool *Pool, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		pool: pool,
		conn: conn,
		send: make(chan []byte),
	}

	pool.register <- client.conn

	go client.Read()
	go client.Write()
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

func setupRoutes(pool *Pool) {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(pool, w, r)
	})
}

func main() {
	fmt.Println("server starting...")
	pool := NewPool()
	go pool.Start()
	setupRoutes(pool)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
