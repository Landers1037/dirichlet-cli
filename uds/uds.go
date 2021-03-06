/*
landers Apps
Author: landers
Github: github.com/landers1037
*/

package uds

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/landers1037/dirichlet_cli/history"
)

// c tcp socket
var c net.Conn

// cmd chan
var cmdChan chan string
var resChan chan string

func init() {
	cmdChan = make(chan string, 1)
	resChan = make(chan string, 1)
}

func Dial() {
	check()
	c, err := net.Dial("unix", FG.Addr)
	history.WriteHist(fmt.Sprintf("try to connect to %s", FG.Addr))
	if err != nil {
		history.WriteHist(fmt.Sprintf("connect to dirichlet failed: %s\n", err.Error()))
		return
	}
	history.WriteHist(fmt.Sprintf("connect to %s success", FG.Addr))
	go func() {
		for {
			select {
			case cmdr := <-cmdChan:
				var userInput = strings.Fields(cmdr)
				if exit(userInput...) {
					c.Close()
					history.WriteHist(ErrExit)
					break
				}

				_, err = c.Write([]byte(strings.Join(userInput, " ")))
				if err != nil {
					history.WriteHist(ErrExit + err.Error())
					break
				}

				var buf = make([]byte, 4096*64)
				cnt, err := c.Read(buf)
				if err != nil {
					c.Close()
					history.WriteHist(ErrExit + err.Error())
					break
				}
				reformatResult(buf[:cnt])

			}
		}
	}()
}

func check() {
	if _, err := os.Stat(FG.Addr); os.IsNotExist(err) {
		history.WriteHist(ErrAddr)
	}
}

func exit(s ...string) bool {
	if len(s) <= 0 {
		return false
	}
	if s[0] == "q" || s[0] == "exit" || s[0] == "\\q" {
		return true
	}
	return false
}

type UDSRes struct {
	Error string `json:"error"`
	Data  string `json:"data"`
}

// 格式化结果到历史记录
func reformatResult(buf []byte) {
	// 只记录错误日志
	var d UDSRes
	err := json.Unmarshal(buf, &d)
	if err != nil {
		resChan <- ErrResponse
	} else {
		// 错误存在时 先展示错误
		if d.Error != "" {
			history.WriteHist(d.Error)
			resChan <- fmt.Sprintf("[ERROR]: %s\n%s", d.Error, d.Data)
		} else {
			resChan <- d.Data
		}
	}
}

func DeferExit() {
	if c != nil {
		c.Close()
	}
}
