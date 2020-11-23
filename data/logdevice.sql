-- Table: public.log

-- DROP TABLE public.log;
DROP table if exists public.logdevice; 
CREATE TABLE public.logdevice
(
    tm timestamp with time zone NOT NULL,
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
COMMENT ON COLUMN public.logdevice.tm
    IS 'Отметка времени';
COMMENT ON COLUMN public.logdevice.id
    IS 'Код устройства';
COMMENT ON COLUMN public.logdevice.crossinfo
    IS 'Привязка к перекрестку';
COMMENT ON COLUMN public.logdevice.txt
    IS 'Сообщение';
