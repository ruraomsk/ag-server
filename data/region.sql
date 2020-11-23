DROP table if exists public.region;
CREATE TABLE public.region
(
    region integer NOT NULL,
    area integer NOT NULL,
    nameregion text COLLATE pg_catalog     ."default" NOT NULL,
    namearea text COLLATE pg_catalog     ."default" NOT NULL

)
    WITH
    (
    OIDS = FALSE
)
TABLESPACE pg_default;

    ALTER TABLE public.region
    OWNER to postgres;

COMMENT ON TABLE public.region
    IS 'Справочник регионов';
COMMENT ON COLUMN public.region.region
    IS 'Идентификатор региона';
COMMENT ON COLUMN public.region.area
    IS 'Идентификатор района';
COMMENT ON COLUMN public.region.nameregion
    IS 'Наименование региона';
COMMENT ON COLUMN public.region.namearea
    IS 'Наименование района';
