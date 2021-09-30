package websocket

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/text/encoding/charmap"
	"io/ioutil"
	"net/url"
	"sync"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
	mu   sync.Mutex
}

type respMsg struct {
	Type int    `json:"type"`
	Id   string `json:"id"`
	Name string `json:"username"`
	Data string `json:"data"`
}

type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		messageType, p, err := c.Conn.ReadMessage()

		outmsg, err := ungzip(p)
		var message Message
		if string(outmsg) == "rub" {
			message = Message{Type: messageType, Body: string("rub")}
		} else {
			var msgJson map[string]interface{}
			err = json.Unmarshal(outmsg, &msgJson)
			if err != nil {
				return
			}

			if msgJson["t"] == "mv" {
				fmt.Println(string(outmsg))

				msgObj := &respMsg{Type: 3, Id: c.ID, Name: c.ID, Data: string(outmsg)}
				newMsg, err := json.Marshal(msgObj)
				if err != nil {
					fmt.Println(err)
				}
				message = Message{Type: messageType, Body: string(newMsg)}
			} else if !(msgJson["t"] == "shs") {
				fmt.Println(string(outmsg))
				returnMsg := &respMsg{
					2, c.ID, c.ID, string(outmsg),
				}
				//message = returnMsg
				newMsg, _ := json.Marshal(returnMsg)
				message = Message{Type: messageType, Body: string(newMsg)}
			}
		}

		c.Pool.Broadcast <- message

	}
}

func ungzip(gzipmsg []byte) (reqmsg []byte, err error) {
	if len(gzipmsg) == 0 {
		return
	}
	if string(gzipmsg) == "rub" {
		reqmsg = gzipmsg
		return
	}
	e := charmap.ISO8859_1.NewEncoder()
	encodeMsg, err := e.Bytes(gzipmsg)
	if err != nil {
		return
	}
	b := bytes.NewReader(encodeMsg)
	r, err := gzip.NewReader(b)
	if err != nil {
		return
	}
	defer r.Close()
	reqmsg, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}
	reqStr, err := url.QueryUnescape(string(reqmsg))
	if err != nil {
		return
	}
	reqmsg = []byte(reqStr)
	return
}
