package views

import (
	"log"
	"websocket-terminal-server/connections"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func SshdSocket(c *gin.Context) {
	var err error
	msg := c.DefaultQuery("msg", "")
	wsConn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Error(err)
		return
	}

	sshClient, err := connections.DecodedMsgToSSHClient(msg)
	if err != nil {
		c.Error(err)
		return
	}
	if sshClient.Host == "" {
		wsConn.WriteMessage(websocket.BinaryMessage, []byte("host can't be null"))
		wsConn.Close()
		return
	}

	if sshClient.Password == "" {
		message := "password can't be null"
		wsConn.WriteMessage(websocket.BinaryMessage, []byte(message))
		wsConn.Close()
		return
	}

	err = sshClient.GenerateClient()
	if err != nil {
		wsConn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
		wsConn.Close()
		return
	}
	client := sshClient.Client
	session, err := client.NewSession()
	if err != nil {
		log.Println(err.Error())
		wsConn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
		wsConn.Close()
		return
	}
	log.Println(session)
	channel, inRequests, err := client.OpenChannel("session", nil)
	if err != nil {
		log.Println("create ssh channel fail", err)
		wsConn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
		wsConn.Close()
		return
	}
	//log.Println("create ssh channel success:", channel)
	go func() {
		for req := range inRequests {
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}()
	//这里第一个协程获取用户的输入
	go func() {
		for {
			// bytes为用户输入
			_, bytes, err := wsConn.ReadMessage()
			if err != nil {
				log.Println(err.Error())
				wsConn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
				wsConn.Close()
				return
			}
			// The first byte is the type of the message.
			if bytes[0] == 0x01 {
				// Make sure bytes is long enough to contain the type length and the type.
				typeLen := int(bytes[1])
				// The next typeLen bytes are req.Type.
				reqType := string(bytes[2 : typeLen+2])

				// The remaining bytes are req.Payload.
				reqPayload := bytes[typeLen+2:]

				// log.Println("received:", reqType, reqPayload)

				if channel == nil {
					msg := "Channel is nil before SendRequest"
					log.Println(msg)
					wsConn.WriteMessage(websocket.BinaryMessage, []byte(msg))
					wsConn.Close()
					return
				} else {
					ok, err := channel.SendRequest(reqType, true, reqPayload)
					if !ok || err != nil {
						log.Println(err.Error())
						wsConn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
						wsConn.Close()
						return

					}
				}
			} else {
				_, err = channel.Write(bytes[1:])
				log.Printf("received input %s\n", string(bytes[1:]))
				if err != nil {
					log.Println(err.Error())
					wsConn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
					wsConn.Close()
					return
				}
			}
		}
	}()

	//第二个协程将远程主机的返回结果返回给用户
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := channel.Read(buf)
			if err != nil {
				wsConn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
				wsConn.Close()
				return
			}
			if n > 0 {
				log.Print("write message:", n)
				wsConn.WriteMessage(websocket.BinaryMessage, buf[:n])
			}
		}
	}()
}
