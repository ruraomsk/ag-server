package gui

import (
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
	"net/http"
	"strconv"
)

func sending(w http.ResponseWriter, res []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(res)
}

func list(w http.ResponseWriter, r *http.Request) {

	res, err := getList()
	if err != nil {
		logger.Error.Println("Запрос ", err.Error())
		return
	}
	sending(w, res)
}
func getlog(w http.ResponseWriter, r *http.Request) {
	ids := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(ids)
	res, err := getLog(id)
	if err != nil {
		logger.Error.Println("Запрос ", err.Error())
		return
	}
	sending(w, res)
}
func getdevice(w http.ResponseWriter, r *http.Request) {
	ids := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(ids)
	res, err := getDevice(id)
	if err != nil {
		logger.Error.Println("Запрос ", err.Error())
		return
	}
	sending(w, res)
}

//Start ответы на запросы от программы визуализации
func Start(context *extcon.ExtContext) {
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/list", list)
	http.HandleFunc("/l", getlog)
	http.HandleFunc("/device", getdevice)
	logger.Info.Println("Listering on port " + strconv.Itoa(setup.Set.Controller.GuiPort))
	err := http.ListenAndServe(":"+strconv.Itoa(setup.Set.Controller.GuiPort), nil)
	if err != nil {
		panic(err.Error())
	}

	// Создаем каналы и переменные
	select {
	case <-context.Done():
		return
	}

}
