package sqlsave

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net"
	"strings"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/secret"
	"github.com/ruraomsk/ag-server/setup"
)

var soc net.Conn
var errSoc error
var connected bool

func sender() bool {
	buffer, err := ioutil.ReadFile(setup.Set.Saver.File)
	if err != nil {
		logger.Error.Printf("Error open file %s", err.Error())
		return false
	}
	lines := bytes.Split(buffer, []byte("\n"))

	//file, err := os.Open(setup.Set.Saver.File)
	//if err != nil {
	//	logger.Error.Printf("Error open file %s", err.Error())
	//	return false
	//}
	//defer file.Close()
	//stat, err := file.Stat()
	//if err != nil {
	//	logger.Error.Printf("Error status file %s", err.Error())
	//	return false
	//}
	if !connected {
		soc, errSoc = net.Dial("tcp", setup.Set.Saver.Remote)
		if errSoc != nil {
			logger.Error.Printf("Error dial %s %s", setup.Set.Saver.Remote, errSoc.Error())
			return false
		}
		connected = true
	}
	if len(lines) == 0 {
		//Send a keep alive
		_, _ = soc.Write([]byte("0\n"))

		return true
	}
	//scanner := bufio.NewScanner(file)
	//buf := make([]byte, 10485760)
	//scanner.Buffer(buf, len(buf))
	reader := bufio.NewReader(soc)
	writer := bufio.NewWriter(soc)
	for _, l := range lines {
		//_ = soc.SetWriteDeadline(time.Now().Add(time.Duration(360 * time.Second)))
		if len(l) < 5 {
			continue
		}
		if setup.Set.Secret {
			_, _ = writer.WriteString(secret.CodeString("==RESPONSE NEED==" + string(l)))
		} else {
			_, _ = writer.WriteString("==RESPONSE NEED==" + string(l))
		}
		_, _ = writer.WriteString("\n")
		errSoc = writer.Flush()
		if errSoc != nil {
			logger.Error.Printf("Error send data %s %s", string(l), errSoc.Error())
			soc.Close()
			connected = false
			return false
		}
		//_ = soc.SetReadDeadline(time.Now().Add(time.Duration(360 * time.Second)))
		response, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Printf("Error read response %s", err.Error())
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
	//if err := scanner.Err(); err != nil {
	//	logger.Error.Printf("Error reading file %s", err.Error())
	//	return false
	//}
	//Coda
	return true
}
