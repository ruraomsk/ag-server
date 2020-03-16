-- Table: public.log

-- DROP TABLE public.log;
DROP table if exists public.log; 
CREATE TABLE public.log
(
    tm timestamp without time zone primary key NOT NULL,
    id integer NOT NULL,
    txt jsonb NOT NULL
)
WITH (
    OIDS = FALSE,
    autovacuum_enabled = TRUE
)
TABLESPACE pg_default;

ALTER TABLE public.log
    OWNER to postgres;
COMMENT ON TABLE public.log
    IS 'Логирование изменения состояния и команд';