-- Table: public."xctrl"

-- DROP TABLE public."xctrl";
DROP table if exists public."xctrl";
CREATE TABLE public."xctrl"
(
    region integer NOT NULL,
    area integer NOT NULL,
    subarea integer NOT NULL,
    switch  boolean ,
	ltime timestamp with time zone , 
	pknow integer,
	pklast integer,
	xnum integer,
    strat jsonb not null,
    calc jsonb not null

)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public."xctrl"
    OWNER to postgres;
