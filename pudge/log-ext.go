package pudge

import (
	"fmt"
)

func (cc *Controller) getPhaseDU() string {
	// return fmt.Sprintf("%d %d %d", cc.DK.FDK, cc.DK.FTUDK, cc.DK.FTSDK)
	switch cc.DK.FTUDK {
	case 0:
		return "ЛР"
	case 9:
		return "ПрТакт"
	case 10:
		return "ЖМ"
	case 11:
		return "ОС"
	case 12:
		return "КК"
	case 14:
		return "ЖМ"
	case 15:
		return "ОС"
	}
	return fmt.Sprintf("%d", cc.DK.FTUDK)
}
func (cc *Controller) getPhaseRU() string {
	// return fmt.Sprintf("%d %d %d", cc.DK.FDK, cc.DK.FTUDK, cc.DK.FTSDK)
	switch cc.DK.FDK {
	case 0:
		return "ОС"
	case 9:
		return "ПрТакт"
	case 10:
		return "ЖМ"
	case 11:
		return "ОС"
	case 12:
		return "КК"
	case 14:
		return "ЖМ"
	case 15:
		return "ОС"
	}
	return fmt.Sprintf("%d", cc.DK.FDK)
}
func (cc *Controller) getRezim() string {
	switch cc.DK.RDK {
	case 1:
		return "РУ"
	case 2:
		return "РП"
	case 3:
		return "ЗУ"
	case 4:
		return "ДУ"
	case 5:
		return "ЛУ"
	case 6:
		return "ЛУ"
	case 7:
		return "РП"
	case 8:
		return "КУ"
	case 9:
		return "КУ"
	}
	return fmt.Sprintf("Режим %d ", cc.DK.RDK)
}
func (cc *Controller) getBroken() string {
	switch cc.DK.EDK {
	case 0:
		return "НОРМ"
	case 1:
		return "ПЕРЕХОД"
	case 2:
		return "ОБРЫВ ЛС"
	case 3:
		return "НГ паритет"
	case 4:
		return "Нет кода"
	case 5:
		return "ОС КОНФЛИКТ"
	case 6:
		return "ЖМ перегорание"
	case 7:
		return "Невкл в коорд"
	case 8:
		return "Неподчинение"
	case 9:
		return "Длинный промтакт"
	case 10:
		return "Нет фазы"
	case 11:
		return "Обрыв ЛС с КЗЦ"
	case 12:
		return "Обрыв ЛС с ЭВМ"
	case 13:
		return "Нет информации"
	}
	return fmt.Sprintf("Неисправность %d ", cc.DK.EDK)
}
func (cc *Controller) getTechRezim() string {
	switch cc.TechMode {
	case 1:
		return "ВР-СК"
	case 2:
		return "ВР-НК"
	case 3:
		return "ДУ-СК"
	case 4:
		return "ДУ-НК"
	case 5:
		return "ДУ-ПК"
	case 6:
		return "РП"
	case 7:
		return "КП ИП"
	case 8:
		return "КП С"
	case 9:
		return "ВР"
	case 10:
		return "ПК ХТ"
	case 11:
		return "ПК КТ"
	case 12:
		return "ПЗУ"
	}
	return fmt.Sprintf("Режим Тех %d ", cc.TechMode)
}
func SetDeviceStatus(id int) (j Journal) {
	cc, _ := GetController(id)
	j.Device = cc.GetSource()
	j.Rezim = cc.getRezim()
	j.Phase = cc.getPhaseDU()
	if j.Rezim == "КУ" {
		if !(j.Phase == "ЛР" || j.Phase == "ЖМ" || j.Phase == "ОС" || j.Phase == "КК") {
			j.Phase = ""
		}
	}
	if j.Rezim == "РУ" {
		j.Phase = cc.getPhaseRU()
	}
	// j.Phase += fmt.Sprintf(" %d %d %d", cc.DK.FDK, cc.DK.FTUDK, cc.DK.FTSDK)
	j.Status = cc.getBroken()
	if (cc.DK.RDK == 9 || cc.DK.RDK == 4 || cc.DK.RDK == 6) && cc.DK.FDK == 0 {
		j.Phase = ""
		j.Rezim = "ЛУ"
	}
	if (cc.DK.RDK == 9 || cc.DK.RDK == 4 || cc.DK.RDK == 6) && cc.DK.FDK == 12 {
		j.Phase = "КК"
	}
	// if j.Device=="ЭВМ" &&
	return
}
func UserDeviceStatus(arm string, command int, param int) (j Journal) {
	// command =-2 выключение сервера
	// command =-3 отключение устройства
	// command =-4 подключение устройства
	// command =-5 авария 220
	// command =-1 привязка
	switch command {
	case -1:
		j.Rezim = "Привязка"
		j.Arm = arm
	case -2:
		j.Rezim = "Останов сервера"
	case -3:
		j.Rezim = "Отключение устройства"
		j.Device = arm
		j.Arm = ""
	case -4:
		j.Rezim = "Подключено устройство"
		j.Arm = arm
	case -5:
		j.Rezim = "Авария 220V"
		j.Device = arm
		j.Arm = ""
	case 9:
		j.Device = "ЭВМ"
		j.Arm = arm
		j.Rezim = "ДУ"
		switch param {
		case 0:
			//Локальный режим
			j.Phase = "ЛР"
		case 9:
			j.Rezim = "КУ"
		case 10:
			j.Phase = "ЖМ"
		case 11:
			j.Phase = "ОС"
		default:
			j.Rezim = "ДУ"
			j.Phase = fmt.Sprintf("%d", param)
		}

	}
	return
}
func SetTechStatus(id int) (j Journal) {
	cc, _ := GetController(id)
	if cc.DK.RDK == 4 || cc.DK.RDK == 8 || cc.DK.RDK == 9 {
		j.Rezim = cc.getTechRezim()
		if cc.StatusCommandDU.IsDUDK1 || cc.StatusCommandDU.IsDUDK2 {
			j.Rezim += " ДУ"
		}
		j.PK = fmt.Sprintf("%d ПК", cc.PK)
		if cc.StatusCommandDU.IsPK {
			j.PK += " ДУ"
		}
		j.CK = fmt.Sprintf("%d CК", cc.CK)
		if cc.StatusCommandDU.IsCK {
			j.CK += " ДУ"
		}
		j.NK = fmt.Sprintf("%d НК", cc.NK)
		if cc.StatusCommandDU.IsNK {
			j.NK += " ДУ"
		}
	} else {
		j.Rezim = cc.getTechRezim()
	}
	j.Status = cc.getBroken()
	return j
}
func UserTechStatus(id int, arm string, command int, param int) (j Journal) {
	cc, _ := GetController(id)
	j.Rezim = cc.getTechRezim()
	j.Arm = arm
	switch command {
	case -2:
		j.Rezim = "Останов сервера"
	case 5:
		j.PK = fmt.Sprintf("%d ПК", param)
	case 6:
		j.CK = fmt.Sprintf("%d СК", param)
	case 7:
		j.NK = fmt.Sprintf("%d HК", param)
	}
	return j
}
func (lr *LogRecord) GetText() string {
	res := ""
	if lr.Type == 0 {
		//Сообщения технологии
		res = fmt.Sprintf("Режим %s %s %s %s %s", lr.Journal.Rezim, lr.Journal.PK, lr.Journal.CK, lr.Journal.NK, lr.Journal.Arm)
	} else {
		res = fmt.Sprintf("%s %s %s %s %s", lr.Journal.Device, lr.Journal.Arm, lr.Journal.Status, lr.Journal.Rezim, lr.Journal.Phase)
	}
	return res
}
