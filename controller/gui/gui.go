package gui

import (
	"net/http"
	"rura/ag-server/extcon"
	"rura/ag-server/setup"
	"rura/teprol/logger"
	"time"
)

func sending(w http.ResponseWriter, res []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(res)
}

func list(w http.ResponseWriter, r *http.Request) {

	res, err := 
	if err != nil {
		logger.Error.Println("Запрос ", err.Error())
		return
	}
	sending(w, res)
}

//Start ответы на запросы от программы визуализации
func Start(context *extcon.ExtContext, stop chan int) {
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/list", list)
	logger.Info.Println("Listering on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err.Error())
	}

	// Создаем каналы и переменные
	context.SetTimeOut(time.Duration(setup.Set.Controller.Step) * time.Second)
	for true {
		select {
		case <-context.Done():
			if context.GetStatus() == "timeout" {
				context.SetTimeOut(time.Duration(setup.Set.Controller.Step) * time.Second)
			} else {
				context.Cancel()
				return
			}
		}
	}

}
