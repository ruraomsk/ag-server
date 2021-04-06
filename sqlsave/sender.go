package sqlsave

import (
	"bufio"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/setup"
	"net"
	"os"
	"strings"
	"time"
)

var soc net.Conn
var errSoc error
var connected bool

func sender() bool {
	file, err := os.Open(setup.Set.Saver.File)
	if err != nil {
		logger.Error.Printf("Error open file %s", err.Error())
		return false
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		logger.Error.Printf("Error status file %s", err.Error())
		return false
	}
	if !connected {
		soc, errSoc = net.Dial("tcp", setup.Set.Saver.Remote)
		if errSoc != nil {
			logger.Error.Printf("Error dial %s %s", setup.Set.Saver.Remote, errSoc.Error())
			return false
		}
		connected = true
	}
	if stat.Size() == 0 {
		//Send a keep alive
		_, _ = soc.Write([]byte("0\n"))
		return true
	}
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 256*1024)
	scanner.Buffer(buf, 256*1024)
	reader := bufio.NewReader(soc)
	writer := bufio.NewWriter(soc)
	for scanner.Scan() {
		_ = soc.SetWriteDeadline(time.Now().Add(time.Duration(10 * int64(time.Second))))
		_, _ = writer.WriteString("==RESPONSE NEED==")
		_, _ = writer.WriteString(scanner.Text())
		_, _ = writer.WriteString("\n")
		errSoc = writer.Flush()
		if errSoc != nil {
			logger.Error.Printf("Error send data %s %s", scanner.Text(), errSoc.Error())
			soc.Close()
			connected = false
			return false
		}
		response, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Printf("Error read response %s", errSoc.Error())
			soc.Close()
			connected = false
			return false
		}
		if strings.Compare(response, "ok\n") != 0 {
			logger.Error.Printf("Response from remote %s", response)
			soc.Close()
			connected = false
			return false
		}

	}
	if err := scanner.Err(); err != nil {
		logger.Error.Printf("Error reading file %s", err.Error())
		return false
	}
	//Coda
	return true
}
