package pudge

import "strconv"

//ToList для вывода в гуи
func (d *DK) ToList(result []string) []string {
	r := "Режим ДК;" + strconv.Itoa(d.RDK)
	result = append(result, r)
	r = "Фаза ДК;" + strconv.Itoa(d.FDK)
	result = append(result, r)
	r = "Устройство ДК;" + strconv.Itoa(d.DDK)
	result = append(result, r)
	r = "Код неисправности ДК;" + strconv.Itoa(d.EDK)
	result = append(result, r)
	r = "Признак переходного периода ДК;" + strconv.FormatBool(d.PDK)
	result = append(result, r)
	r = "Дополнительный код неисправности ДК;" + strconv.Itoa(d.EEDK)
	result = append(result, r)
	r = "Открыта дверь ДК;" + strconv.FormatBool(d.ODK)
	result = append(result, r)
	r = "Номер фазы на которой сгорели лампы ДК;" + strconv.Itoa(d.LDK)
	result = append(result, r)
	r = "Фаза ТУ ДК на момент передачи;" + strconv.Itoa(d.FTUDK)
	result = append(result, r)
	r = "Время отработки ТУ в секундах;" + strconv.Itoa(d.TDK)
	result = append(result, r)
	r = "Фаза ТС ДК;" + strconv.Itoa(d.FTSDK)
	result = append(result, r)
	r = "Время от начала фазы в секундах;" + strconv.Itoa(d.TTCDK)
	result = append(result, r)
	return result
}

//ToList для вывода в гуи
func (m *Model) ToList(result []string) []string {
	r := "Версия ПО пдаты ПСПД;" + strconv.Itoa(m.VPCPD)
	result = append(result, r)
	r = "Версия ПО пдаты ПБС;" + strconv.Itoa(m.VPBS)
	result = append(result, r)
	r = "Субблок С12;" + strconv.FormatBool(m.C12)
	result = append(result, r)
	r = "Разрешение накопления статистики по ТП;" + strconv.FormatBool(m.STP)
	result = append(result, r)
	r = "Контроллер ДК-А;" + strconv.FormatBool(m.DKA)
	result = append(result, r)
	r = "Детектор транспорта;" + strconv.FormatBool(m.DTA)
	result = append(result, r)
	return result
}

//ToList для вывода в гуи
func (e *ErrorDevice) ToList(result []string) []string {
	r := "Срабатывание входа контроля 220В DK1;" + strconv.FormatBool(e.V220DK1)
	result = append(result, r)
	r = "Срабатывание входа контроля 220В DK2;" + strconv.FormatBool(e.V220DK2)
	result = append(result, r)
	r = "Неисправность часов RTC;" + strconv.FormatBool(e.RTC)
	result = append(result, r)
	r = "Неисправность ТВП1;" + strconv.FormatBool(e.TVP1)
	result = append(result, r)
	r = "Неисправность ТВП2;" + strconv.FormatBool(e.TVP2)
	result = append(result, r)
	r = "Неисправность FRAM;" + strconv.FormatBool(e.FRAM)
	result = append(result, r)
	return result
}

//ToList для вывода в гуи
func (g *GPS) ToList(result []string) []string {
	r := "Исправно;" + strconv.FormatBool(g.Ok)
	result = append(result, r)
	r = "Нет связи с приемником;" + strconv.FormatBool(g.E01)
	result = append(result, r)
	r = "Ошибка CRC;" + strconv.FormatBool(g.E02)
	result = append(result, r)
	r = "Нет валидного времени;" + strconv.FormatBool(g.E03)
	result = append(result, r)
	r = "Мало спутников;" + strconv.FormatBool(g.E04)
	result = append(result, r)
	r = "Поиск спутников после включения;" + strconv.FormatBool(g.Seek)
	return result
}

//ToList для вывода в гуи
func (i *Input) ToList(result []string) []string {
	r := "Неисправность входа 1;" + strconv.FormatBool(i.V1)
	result = append(result, r)
	r = "Неисправность входа 2;" + strconv.FormatBool(i.V2)
	result = append(result, r)
	r = "Неисправность входа 3;" + strconv.FormatBool(i.V3)
	result = append(result, r)
	r = "Неисправность входа 4;" + strconv.FormatBool(i.V4)
	result = append(result, r)
	r = "Неисправность входа 5;" + strconv.FormatBool(i.V5)
	result = append(result, r)
	r = "Неисправность входа 6;" + strconv.FormatBool(i.V6)
	result = append(result, r)
	r = "Неисправность входа 7;" + strconv.FormatBool(i.V7)
	result = append(result, r)
	r = "Неисправность входа 8;" + strconv.FormatBool(i.V8)
	result = append(result, r)
	return result
}

//ToList для вывода в гуи
func (s *StatusCommandDU) ToList(result []string) []string {
	r := "Назначен ПК;" + strconv.FormatBool(s.IsPK)
	result = append(result, r)
	r = "Назначена карта выбора по времени суток;" + strconv.FormatBool(s.IsCK)
	result = append(result, r)
	r = "Назначена недельная карта;" + strconv.FormatBool(s.IsNK)
	result = append(result, r)
	r = "На  ДК есть команда ДУ;" + strconv.FormatBool(s.IsDUDK1)
	result = append(result, r)
	r = "Есть запрос на передачу фаз по ДК СФДК;" + strconv.FormatBool(s.IsReqSFDK1)
	result = append(result, r)
	return result
}

//ToList для вывода в гуи
func (c *Controller) ToList() []string {
	result := make([]string, 0)
	r := "Технологический режим;" + strconv.Itoa(c.TexRezim)
	result = append(result, r)
	r = "Работаем в базовой привязке;" + strconv.FormatBool(c.Base)
	result = append(result, r)
	r = "Номер плана координации;" + strconv.Itoa(c.PK)
	result = append(result, r)
	r = "Номер суточной карты;" + strconv.Itoa(c.CK)
	result = append(result, r)
	r = "Номер недельной карты;" + strconv.Itoa(c.NK)
	result = append(result, r)
	r = "Максимальное время ожидания от сервера;" + strconv.Itoa(int(c.TMax))
	result = append(result, r)
	r = "ДК; "
	result = append(result, r)
	result = c.DK.ToList(result)
	r = "МОДЕЛЬ УСТРОЙСТВА; "
	result = append(result, r)
	result = c.Model.ToList(result)
	r = "ОШИБКИ УСТРОЙСТВА; "
	result = append(result, r)
	result = c.Error.ToList(result)
	r = "СОСТОЯНИЕ GPS ПРИЕМНИКА; "
	result = append(result, r)
	result = c.GPS.ToList(result)
	r = "ТЕКУЩИЕ КОМАНДЫ ДУ; "
	result = append(result, r)
	result = c.StatusCommandDU.ToList(result)
	r = "ТЕКУЩИЕ НЕИСПРАВНОСТИ ВХОДОВ; "
	result = append(result, r)
	result = c.Input.ToList(result)
	return result
}
