package creator

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"time"

	//Инициализатор постргресса
	_ "github.com/lib/pq"

	"github.com/ruraomsk/ag-server/creator/svg"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
)

var dbb *sql.DB
var err error
var create = `
-- Table: public.svg

-- DROP TABLE IF EXISTS public.svg;

CREATE TABLE IF NOT EXISTS public.svg
(
    region integer NOT NULL,
    area integer NOT NULL,
    id integer NOT NULL,
    state jsonb NOT NULL DEFAULT '{}'::jsonb,
    bottom bytea,
    picture bytea,
    extend jsonb NOT NULL DEFAULT '{}'::jsonb
)
WITH (
    OIDS = FALSE,
    autovacuum_enabled = TRUE
)
TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.svg
    OWNER to postgres;

COMMENT ON TABLE public.svg
    IS 'Таблица для проверки bytea';
`

func SvgCreator() error {
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	for {
		dbb, err = sql.Open("postgres", dbinfo)

		if err != nil {
			fmt.Println(err.Error())
			logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
			time.Sleep(time.Second * 10)
			continue
		}
		if err = dbb.Ping(); err != nil {
			logger.Error.Printf("Ping %s", err.Error())
			time.Sleep(time.Second * 10)
			continue
		}
		break

	}
	defer dbb.Close()
	fmt.Println("SVG creator start")
	logger.Info.Println("SVG creator start")
	dbb.Exec(create)
	w := "select region,area,id from public.\"cross\";"
	crs, err := dbb.Query(w)
	if err != nil {
		logger.Error.Printf("%s %s ", w, err.Error())
		return err
	}
	count := 0
	for crs.Next() {
		var region, area, id int
		crs.Scan(&region, &area, &id)
		count++
		fmt.Print(".")
		if count%50 == 0 {
			fmt.Println()
		}
		logger.Info.Printf("load %d %d %d", region, area, id)
		picture, err := ioutil.ReadFile(getPath(region, area, id) + "cross.svg")
		if err != nil {
			logger.Error.Printf(" %s ", err.Error())
			continue
		}
		extend, err := svg.Parse(getPath(region, area, id) + "cross.svg")
		if err != nil {
			logger.Error.Printf(" %s ", err.Error())
			continue
		}

		bottom, err := ioutil.ReadFile(getPath(region, area, id) + "map.png")
		if err != nil {
			logger.Error.Printf(" %s ", err.Error())
			continue
		}
		state, err := ioutil.ReadFile(getPath(region, area, id) + "template.tmpl")
		if err != nil {
			state = []byte("{}")
		}
		w := fmt.Sprintf("select count(*) from public.svg where region=%d and area=%d and id=%d;", region, area, id)
		svgs, err := dbb.Query(w)
		if err != nil {
			logger.Error.Printf("%s %s ", w, err.Error())
			continue
		}
		found := false
		for svgs.Next() {
			var count int
			svgs.Scan(&count)
			if count > 0 {
				found = true
			}
		}
		svgs.Close()

		if found {
			w = fmt.Sprintf("update public.svg set state=$1,bottom=$2,picture=$3,extend=$4 where  region=%d and area=%d and id=%d;", region, area, id)
			_, err = dbb.Exec(w, state, bottom, picture, extend)
		} else {
			w = "INSERT INTO public.svg(region, area, id, state, bottom, picture,extend) VALUES ($1, $2, $3, $4, $5, $6,$7);"
			_, err = dbb.Exec(w, region, area, id, state, bottom, picture, extend)
		}
		if err != nil {
			logger.Error.Printf("%s %s", w, err.Error())
		}
		fmt.Print("*")

	}
	crs.Close()
	fmt.Print("\nStart to verify\n")
	cmap := make(map[string]bool)
	w = "SELECT region, area, id, bottom, picture 	FROM public.svg;"
	svgs, err := dbb.Query(w)
	if err != nil {
		logger.Error.Printf("%s %s ", w, err.Error())
		return err
	}

	count = 0
	for svgs.Next() {
		count++
		fmt.Print(".")
		if count%50 == 0 {
			fmt.Println()
		}

		var region, area, id int
		var bottom, picture []byte
		svgs.Scan(&region, &area, &id, &bottom, &picture)
		logger.Info.Printf("verify %d %d %d", region, area, id)
		cmap[fmt.Sprintf("%d/%d/%d", region, area, id)] = true
		dpicture, err := ioutil.ReadFile(getPath(region, area, id) + "cross.svg")
		if err != nil {
			logger.Error.Printf(" %s ", err.Error())
			continue
		}
		dbottom, err := ioutil.ReadFile(getPath(region, area, id) + "map.png")
		if err != nil {
			logger.Error.Printf(" %s ", err.Error())
			continue
		}
		// dstate, err := ioutil.ReadFile(getPath(region, area, id) + "template.tmpl")
		// if err != nil {
		// 	logger.Error.Printf(" %s ", err.Error())
		// 	dstate = []byte("{}")
		// }
		if !reflect.DeepEqual(&picture, &dpicture) {
			logger.Error.Printf("cross,svg not equal")
		}
		if !reflect.DeepEqual(&bottom, &dbottom) {
			logger.Error.Printf("map.png not equal")
		}
		// if !reflect.DeepEqual(&state, &dstate) {
		// 	logger.Error.Printf("template.tmpl not equal")
		// }
		fmt.Print("*")
	}
	fmt.Print("\nStart to test time\n")
	start := time.Now()
	count = 0
	for k := range cmap {
		count++
		fmt.Print(".")
		if count%50 == 0 {
			fmt.Println()
		}
		ks := strings.Split(k, "/")
		rows, err := dbb.Query("select picture from public.svg where region=$1 and area=$2 and id=$3;", ks[0], ks[1], ks[2])
		if err != nil {
			logger.Error.Printf(" %s ", err.Error())
			continue
		}
		for rows.Next() {
			var picture []byte
			rows.Scan(&picture)
			fmt.Print("*")

		}
	}
	l := time.Since(start).Milliseconds()
	fmt.Printf("\nmsec/ops %d %d\n", l, count)
	fmt.Println("SVG creator end")
	logger.Info.Println("SVG creator end")
	return nil
}

func getPath(region, area, id int) string {
	return fmt.Sprintf("%s/%d/%d/%d/", setup.Set.Loader.Path, region, area, id)
}
