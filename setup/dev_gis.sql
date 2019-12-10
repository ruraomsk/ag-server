-- Table: public.dev_gis

-- DROP TABLE public.dev_gis;
DROP table if exists public.dev_gis;
CREATE TABLE public.dev_gis
(
    region integer NOT NULL,
    id integer NOT NULL,
    dgis point NOT NULL,
    describ text

)
WITH (
    OIDS = FALSE,
    autovacuum_enabled = TRUE
)
TABLESPACE pg_default;

ALTER TABLE public.dev_gis
    OWNER to postgres;
COMMENT ON COLUMN public.dev_gis.region
    IS 'Идентификатор региона';

COMMENT ON COLUMN public.dev_gis.id
    IS 'Идентификатор перекрестка';

COMMENT ON COLUMN public.dev_gis.dgis
    IS 'Координаты перекрестка';

COMMENT ON COLUMN public.dev_gis.describ
    IS 'Адрес перекрестка';