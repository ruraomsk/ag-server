-- Table: public.log

-- DROP TABLE public.log;
DROP table if exists public.logdevice; 
CREATE TABLE public.logdevice
(
    tm timestamp without time zone primary key NOT NULL,
    id integer NOT NULL,
    crossinfo jsonb,
    txt text NOT NULL
)
WITH (
    OIDS = FALSE,
    autovacuum_enabled = TRUE
)
TABLESPACE pg_default;

ALTER TABLE public.logdevice
    OWNER to postgres;
COMMENT ON TABLE public.logdevice
    IS 'Логирование изменения состояния и команд';