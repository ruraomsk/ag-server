package techComm

import (
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
)

func calcStatus(cc pudge.Controller) int {
	rezim := cc.DK.RDK
	faza := cc.DK.FDK
	err := cc.DK.EDK
	dev := cc.DK.DDK
	lamp := 1
	if cc.DK.LDK == 0 {
		lamp = 0
	}

	door := 1
	if !cc.DK.ODK {
		door = 0
	}

	//	dev := cc.codeDevice()
	// 1 - ВПУ
	// 2 - ДК
	// 3 - ИП УСДК
	// 4 - УСДК
	// 5 - ИП ДК
	//
	// 7 - ДУ ЭВМ

	//rezim := cc.DK.RDK
	// 1 - РУ
	// 2 - РП
	// 3 - ЗУ
	// 4 - ДУ
	// 6 - ЛР
	// 7 - ЛРП
	// 8 - МГР
	// 9 - КУ
	// 10 - РКУ

	// }
	if !cc.StatusConnection {
		if err == 11 && dev == 3 {
			//Авария 220 16
			return 16
		}
		if err == 11 && dev == 5 { //Спросить как узнать ошибки контроллера УСДК
			//Выключен УСДК/ДК 17
			return 17
		}
		return 18
	}
	if cc.Base {
		return 22
	}
	if (rezim == 8 || rezim == 9) && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) && lamp == 0 && door == 0 {
		//Координированное управление 1
		return 1
	}
	if rezim == 4 && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) {
		//Диспетчерское управление 2
		return 2
	}
	if (rezim == 1 || rezim == 2) && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) {
		//Ручное управление 3
		return 3
	}
	if (rezim == 1) && (err == 0 || err == 1) && (faza == 12) {
		//Ручное управление 3
		return 3
	}
	if rezim == 3 && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) && lamp == 0 && door == 0 {
		//Зеленая улица 4
		return 4
	}
	if (rezim == 5 || rezim == 6) && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) && door == 0 {
		//Локальное управление 5
		return 5
	}
	if (rezim == 9 || rezim == 4 || rezim == 6) && (err == 0 || err == 1) && (faza == 0) && door == 0 {
		//Локальное управление 5
		return 5
	}

	if (rezim == 8 || rezim == 9) && (err == 0 || err == 1) && (faza == 10) && door == 0 {
		//Желтое мигание по расписанию 6
		return 6
	}
	if rezim == 4 && (err == 0 || err == 1) && (faza == 10) {
		//Желтое мигание из центра 7
		return 7
	}
	if (rezim == 1 || rezim == 2) && (err == 0 || err == 1) && faza == 10 {
		//Желтое мигание заданное на перекрестке 8
		return 8
	}
	if (rezim == 5 || rezim == 6) && (err == 0 || err == 1) && faza == 10 && door == 0 {
		//Желтое мигание по расписанию 9
		return 9
	}
	if (rezim == 8 || rezim == 9 || rezim == 4) && (err == 0 || err == 1) && faza == 12 && door == 0 {
		//Кругом красный 10
		return 10
	}
	if (rezim == 8 || rezim == 9) && (err == 0 || err == 1) && faza == 11 && door == 0 {
		//Отключение светофора по расписанию 11
		return 11
	}
	if rezim == 4 && (err == 0 || err == 1) && faza == 11 {
		//Желтое мигание заданное из центра 12 1
		return 12
	}
	if (rezim == 1 || rezim == 2) && (err == 0 || err == 1) && faza == 11 {
		//Отключение светофора заданное на перекрестке 13
		return 13
	}
	if (rezim == 5 || rezim == 6) && (err == 0 || err == 1) && faza == 11 && door == 0 {
		//Отключение светофора по расписанию ДК 14
		return 14
	}
	if err == 11 && dev == 3 {
		//Авария 220 16
		return 16
	}
	if err == 11 && dev == 5 { //Спросить как узнать ошибки контроллера УСДК
		//Выключен УСДК/ДК 17
		return 17
	}
	if err == 11 && dev == 4 { //Спросить как узнать ошибки GPRS
		//Нет связи с УСДК 18
		return 18
	}
	if err == 11 && dev == 8 { //Спросить как узнать ошибки ПБС УСДК
		//Нет связи с ПСПД 19
		return 19
	}
	if err == 11 {
		//Обрыв ЛС КЗЦ 20
		return 20
	}
	if err == 4 && dev == 8 { //Спросить как узнать ошибки ПБС УСДК
		//Превышение трафика
		return 21
	}
	if (rezim == 5 || rezim == 6) && err == 4 && (dev == 0 || dev == 1) {
		//Базовая привязка 22
		return 22
	}
	if (rezim == 1 || rezim == 2) && err == 4 && dev == 4 {
		//Неисправность часов или GPS 22
		return 23
	}
	if (rezim == 5 || rezim == 6) && err == 4 && (dev == 4 || dev == 5) {
		//Коррекция привязки 24
		return 24
	}
	if err == 10 {
		//Несуществующая фаза
		return 25
	}
	if err == 4 {
		//Несуществующий код
		return 26
	}
	if (rezim == 8 || rezim == 9) && (faza > 0 && faza < 10) && lamp == 1 {
		//Координированное управление и перегоревшая лампа
		return 27
	}
	if err == 2 {
		//Обрыв линий связи
		return 28
	}
	if err == 3 {
		//Негоден по паритету
		return 29
	}
	if err == 5 && faza == 11 {
		//Отключен из-за конфликта направлений
		return 30
	}
	if err == 5 {
		//Конфликт направлений
		return 31
	}
	if err == 6 && faza == 10 {
		//Желтое мигание из-за перегорания
		return 32
	}
	if err == 6 {
		//Не годен по перегоранию ламп
		return 33
	}
	if err == 7 {
		//Не включается в координацию
		return 34
	}
	if err == 8 {
		//Дорожный контроллер не подчиняется командам
		return 35
	}
	if err == 9 {
		//Длинный промежуточный такт
		return 36
	}
	if err == 12 {
		//Обрыв линий связи ЭВМ с перекрестками
		return 37
	}
	if err == 3 {
		//Нет информации о работе перекрестка
		return 38
	}
	if door != 0 {
		//Двери открыты 15
		return 15
	}
	logger.Debug.Printf("Режим=%v Фаза=%v Ошибка=%v Устройство=%v Лампа=%v Дверь=%v ID %v", rezim, faza, err, dev, lamp, door, cc.ID)
	return 39
}
func calcJournal(cc pudge.Controller) (int, int) {
	rezim := cc.DK.RDK
	faza := cc.DK.FDK
	err := cc.DK.EDK
	dev := cc.DK.DDK
	lamp := 1
	if cc.DK.LDK == 0 {
		lamp = 0
	}
	//lamp := 0
	//door := 0
	// if lrezim != rezim || lfaza != faza || lerr != err || ldev != dev || llamp != lamp || ldoor != door {
	// 	logger.Info.Printf("rezim=%d faza=%d err=%d dev=%d lamp=%d door=%d", rezim, faza, err, dev, lamp, door)
	// 	lrezim = rezim
	// 	lfaza = faza
	// 	ldev = dev
	// 	llamp = lamp
	// 	ldoor = door
	// }
	if !cc.StatusConnection {
		if err == 11 && dev == 3 {
			//Авария 220 16
			return -1, 16
		}
		if err == 11 && dev == 5 { //Спросить как узнать ошибки контроллера УСДК
			//Выключен УСДК/ДК 17
			return -1, 17
		}

		return -1, 18
	}
	if cc.Base {
		return 1, 22
	}
	if (rezim == 8 || rezim == 9) && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) {
		//Координированное управление 1
		return -1, 1
	}
	if rezim == 4 && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) {
		//Диспетчерское управление 2
		return -1, 2
	}
	if (rezim == 1 || rezim == 2) && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) {
		//Ручное управление 3
		return -1, 3
	}
	if rezim == 3 && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) {
		//Зеленая улица 4
		return -1, 4
	}
	if (rezim == 5 || rezim == 6) && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) {
		//Локальное управление 5
		return -1, 5
	}
	if (rezim == 9 || rezim == 4 || rezim == 6) && (err == 0 || err == 1) && (faza == 0) {
		//Локальное управление 5
		return -1, 5
	}
	if (rezim == 8 || rezim == 9) && (err == 0 || err == 1) && (faza == 10) {
		//Желтое мигание по расписанию 6
		return -1, 6
	}
	if rezim == 4 && (err == 0 || err == 1) && (faza == 10) {
		//Желтое мигание из центра 7
		return -1, 7
	}
	if (rezim == 1 || rezim == 2) && (err == 0 || err == 1) && faza == 10 {
		//Желтое мигание заданное на перекрестке 8
		return -1, 8
	}
	if (rezim == 5 || rezim == 6) && (err == 0 || err == 1) && faza == 10 {
		//Желтое мигание по расписанию 9
		return -1, 9
	}
	if (rezim == 8 || rezim == 9 || rezim == 4) && (err == 0 || err == 1) && faza == 12 {
		//Кругом красный 10
		return 1, 10
	}
	if (rezim == 8 || rezim == 9) && (err == 0 || err == 1) && faza == 11 {
		//Отключение светофора по расписанию 11
		return -1, 11
	}
	if rezim == 4 && (err == 0 || err == 1) && faza == 11 {
		//Желтое мигание заданное из центра 12 1
		return -1, 12
	}
	if (rezim == 1 || rezim == 2) && (err == 0 || err == 1) && faza == 11 {
		//Отключение светофора заданное на перекрестке 13
		return -1, 13
	}
	if (rezim == 5 || rezim == 6) && (err == 0 || err == 1) && faza == 11 {
		//Отключение светофора по расписанию ДК 14
		return -1, 14
	}
	if err == 11 && dev == 3 {
		//Авария 220 16
		return -1, 16
	}
	if err == 11 && dev == 5 { //Спросить как узнать ошибки контроллера УСДК
		//Выключен УСДК/ДК 17
		return -1, 17
	}
	if err == 11 && dev == 4 { //Спросить как узнать ошибки GPRS
		//Нет связи с УСДК 18
		return -1, 18 //18
	}
	if err == 11 && dev == 8 { //Спросить как узнать ошибки ПБС УСДК
		//Нет связи с ПСПД 19
		return 1, 19
	}
	if err == 11 {
		//Обрыв ЛС КЗЦ 20
		return 1, 20
	}
	if err == 4 && dev == 8 { //Спросить как узнать ошибки ПБС УСДК
		//Превышение трафика
		return 1, 21
	}
	if (rezim == 5 || rezim == 6) && err == 4 && (dev == 0 || dev == 1) {
		//Базовая привязка 22
		return 1, 22
	}
	if (rezim == 1 || rezim == 2) && err == 4 && dev == 4 {
		//Неисправность часов или GPS 22
		return -1, 23
	}
	if (rezim == 5 || rezim == 6) && err == 4 && (dev == 4 || dev == 5) {
		//Коррекция привязки 24
		return 1, 24
	}
	if err == 10 {
		//Несуществующая фаза
		return -1, 25
	}
	if err == 4 {
		//Несуществующий код
		return -1, 26
	}
	if (rezim == 8 || rezim == 9) && (faza > 0 && faza < 10) && lamp == 1 {
		//Координированное управление и перегоревшая лампа
		return -1, 27
	}
	if err == 2 {
		//Обрыв линий связи
		return -1, 28
	}
	if err == 3 {
		//Негоден по паритету
		return 1, 29
	}
	if err == 5 && faza == 11 {
		//Отключен из-за конфликта направлений
		return -1, 30
	}
	if err == 5 {
		//Конфликт направлений
		return -1, 31
	}
	if err == 6 && faza == 10 {
		//Желтое мигание из-за перегорания
		return -1, 32
	}
	if err == 6 {
		//Не годен по перегоранию ламп
		return -1, 33
	}
	if err == 7 {
		//Не включается в координацию
		return 1, 34
	}
	if err == 8 {
		//Дорожный контроллер не подчиняется командам
		return 1, 35
	}
	if err == 9 {
		//Длинный промежуточный такт
		return 1, 36
	}
	if err == 12 {
		//Обрыв линий связи ЭВМ с перекрестками
		return -1, 37
	}
	if err == 3 {
		//Нет информации о работе перекрестка
		return -1, 38
	}
	logger.Debug.Printf("Режим=%d Фаза=%d Ошибка=%d Устройство=%d ID %d", rezim, faza, err, dev, cc.ID)
	return -1, 39
}
