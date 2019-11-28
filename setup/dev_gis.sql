-- Table: public.dev_gis

-- DROP TABLE public.dev_gis;
DROP table if exists public.dev_gis; 
CREATE TABLE public.dev_gis
(
    id integer primary key NOT NULL,
    dgis point NOT NULL,
    describ text ,

)
WITH (
    OIDS = FALSE,
    autovacuum_enabled = TRUE
)
TABLESPACE pg_default;

ALTER TABLE public.dev_gis
    OWNER to postgres;

COMMENT ON COLUMN public.dev_gis.id
    IS 'Идентификатор устройства';

COMMENT ON COLUMN public.dev_gis.dgis
    IS 'Координаты устройства';

COMMENT ON COLUMN public.dev_gis.describ
    IS 'Адрес устройства';