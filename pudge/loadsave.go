package pudge

import (
	"database/sql"
	"encoding/json"
	"github.com/ruraomsk/TLServer/logger"
	"strconv"
)

func loadDBase() error {
	err := loadControllers()
	if err != nil {
		return err
	}
	//err := clearControllers()
	//if err != nil {
	//	return err
	//}

	err = loadCrosees()
	if err != nil {
		return err
	}
	err = loadStatuses()
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
func loadStatuses() error {
	rows, err := conCross.Query("Select id,description,control from public.status;")
	if err != nil {
		return err
	}
	defer rows.Close()
	var id int
	var description string
	var control bool
	for rows.Next() {
		err = rows.Scan(&id, &description, &control)
		if err != nil {
			return err
		}
		statuses[id] = description
		controls[id] = control
	}
	return nil
}
func loadCrosees() error {
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
		nowstatus[reg.ToKey()] = ""
	}
	return nil
}
func loadControllers() error {
	rows, err := conDBSave.Query("Select id,device from devices;")

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
		c.StatusConnection = false
		c.DK.EDK = 0
		c.DK.PDK = false
		_, is := controllers[id]
		if is {
			logger.Error.Printf("?????????????????????? id  ???????????????????? %d", id)
		}
		controllers[id] = c
	}
	return nil
}
func saveControllers() error {
	// logger.Debug.Println("saveControllers")
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
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
	// logger.Debug.Println("saveCrossers")
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	count := 0
	for _, c := range crosses {
		if !c.WriteToDB {
			continue
		}
		js, _ := json.Marshal(c)
		_, err = conCross.Exec("update  public.cross set idevice=" + strconv.Itoa(c.IDevice) + ",state='" + string(js) + "',describ='" + c.Name + "',dgis='" +
			c.Dgis + "',status=" + strconv.Itoa(c.StatusDevice) + ",subarea=" + strconv.Itoa(c.SubArea) + " where region=" +
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

var historyTable = `
	CREATE TABLE public.history
	(
		region integer NOT NULL,
		area integer NOT NULL,
		id integer NOT NULL,
		login text ,
		tm timestamp with time zone NOT NULL,
		state jsonb NOT NULL
	)
	WITH (
		autovacuum_enabled = TRUE
	)
	TABLESPACE pg_default;
	
	ALTER TABLE public.history
		OWNER to postgres;
`

func needHistoryCross(db *sql.DB) error {
	_, err := db.Exec("select * from public.history;")
	if err != nil {
		logger.Error.Println("history table not found - created!")
		_, _ = db.Exec(historyTable)
	}
	return nil
}
