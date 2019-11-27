-- Table: public.devices

-- DROP TABLE public.devices;
DROP table if exists public.devices; 
CREATE TABLE public.devices
(
    id integer primary key NOT NULL,
    device jsonb NOT NULL
)
WITH (
    OIDS = FALSE,
    autovacuum_enabled = TRUE
)
TABLESPACE pg_default;

ALTER TABLE public.devices
    OWNER to postgres;
COMMENT ON TABLE public.devices
    IS 'Сохранение состояния контроллеров';