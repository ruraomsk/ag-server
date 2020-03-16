package pudge

import (
	"encoding/json"
	"strconv"

	"github.com/JanFant/TLServer/logger"
)

func loadDBase() error {
	err := loadControllers()
	if err != nil {
		return err
	}
	err = loadCrosees()
	if err != nil {
		return err
	}

	return nil
}
func saveDBase() error {
	err := saveControllers()
	if err != nil {
		return err
	}
	err = saveCrosees()
	if err != nil {
		return err
	}

	return nil
}

func loadCrosees() error {
	mutex.Lock()
	defer mutex.Unlock()
	rows, err := conCross.Query("Select region,area,id,idevice,describ,state from public.cross;")
	if err != nil {
		return err
	}
	defer rows.Close()
	var region int
	var area int
	var id int
	var idevice int
	var describ string
	var state []byte
	for rows.Next() {
		c := new(Cross)
		err = rows.Scan(&region, &area, &id, &idevice, &describ, &state)
		if err != nil {
			return err
		}
		err = json.Unmarshal(state, &c)
		if err != nil {
			return err
		}
		c.Region = region
		c.Area = area
		c.ID = id
		c.IDevice = idevice
		c.Name = describ
		c.StatusDevice = 18
		c.WriteToDB = true

		reg := Region{Region: region, Area: area, ID: id}
		crosses[reg.ToKey()] = c
	}
	return nil
}
func loadControllers() error {
	mutex.Lock()
	defer mutex.Unlock()
	rows, err := conDBSave.Query("Select * from devices;")
	var id int
	var js []byte
	for rows.Next() {
		c := new(Controller)
		err = rows.Scan(&id, &js)
		if err != nil {
			return err
		}
		err = json.Unmarshal(js, &c)
		if err != nil {
			return err
		}
		c.WriteToDB = true
		c.StatusConnection = NotConnected

		_, is := controllers[id]
		if is {
			logger.Error.Printf("Дубликатный id  устройства %d", id)
		}
		controllers[id] = c
	}
	return nil
}
func saveControllers() error {
	mutex.Lock()
	defer mutex.Unlock()
	count := 0
	for _, c := range controllers {
		if len(c.Name) == 0 {
			c.Name = getNameCross(c.ID)

			c.WriteToDB = true
		}
		if !c.WriteToDB {
			continue
		}
		count++
		js, _ := json.Marshal(c)
		_, err = conDBSave.Exec("update  devices set device='" + string(js) + "' where id=" + strconv.Itoa(c.ID) + ";")
		if err != nil {
			logger.Error.Printf("For update save to controller %s", err.Error())
			break
		}
		c.WriteToDB = false
		// controllers[c.ID] = c
	}
	// logger.Info.Println("controllers ", count)
	return nil
}
func saveCrosees() error {
	mutex.Lock()
	defer mutex.Unlock()
	count := 0
	for _, c := range crosses {
		if !c.WriteToDB {
			continue
		}
		js, _ := json.Marshal(c)
		_, err = conCross.Exec("update  public.cross set idevice=" + strconv.Itoa(c.IDevice) + ",state='" + string(js) + "',describ='" + c.Name + "',dgis='" + c.Dgis + "' where region=" +
			strconv.Itoa(c.Region) + " and id=" + strconv.Itoa(c.ID) + " and area=" + strconv.Itoa(c.Area) + ";")
		if err != nil {
			logger.Error.Printf("For update save to cross %s", err.Error())
			break
		}
		c.WriteToDB = false
		// reg := Region{Region: c.Region, Area: c.Area, ID: c.ID}
		// crosses[reg.ToKey()] = c
		count++
	}
	// logger.Info.Println("save cross ", count)
	return nil
}
