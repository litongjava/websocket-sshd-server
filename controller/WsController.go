package views

import (
  "github.com/gin-gonic/gin"
  "strconv"
  "websocket-terminal-server/connections"
)

func ShellWs(c *gin.Context) {
  var err error
  msg := c.DefaultQuery("msg", "")
  cols := c.DefaultQuery("cols", "150")
  rows := c.DefaultQuery("rows", "35")
  col, _ := strconv.Atoi(cols)
  row, _ := strconv.Atoi(rows)
  terminal := connections.Terminal{
    Columns: uint32(col),
    Rows:    uint32(row),
  }
  sshClient, err := connections.DecodedMsgToSSHClient(msg)
  if err != nil {
    c.Error(err)
    return
  }
  if sshClient.Host == "" {
    c.Error(&ApiError{Message: "host can't be null", Code: 400})
    return
  }

  if sshClient.Password == "" {
    c.Error(&ApiError{Message: "password can't be null", Code: 400})
    return
  }
  conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
  if err != nil {
    c.Error(err)
    return
  }
  err = sshClient.GenerateClient()
  if err != nil {
    conn.WriteMessage(1, []byte(err.Error()))
    conn.Close()
    return
  }
  sshClient.RequestTerminal(terminal)
  sshClient.Connect(conn)
}
