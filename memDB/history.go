package memDB

import (
	"database/sql"
	"github.com/ruraomsk/TLServer/logger"
)

var historyTable = `
	CREATE TABLE if not exists public.history
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

func needHistoryCross(db *sql.DB) {
	_, err := db.Exec(historyTable)
	if err != nil {
		logger.Error.Println("history table create %s", err.Error())
	}
}
