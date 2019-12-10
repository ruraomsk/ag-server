DROP table if exists public.region;
CREATE TABLE public.region
(
    region integer NOT NULL,
    name text COLLATE pg_catalog
    ."default" NOT NULL
)
    WITH
    (
    OIDS = FALSE
)
TABLESPACE pg_default;

    ALTER TABLE public.region
    OWNER to postgres;

COMMENT ON COLUMN public.region.region
    IS 'Идентификатор региона';

COMMENT ON COLUMN public.region.name
    IS 'Наименование региона';

    INSERT INTO public.region
        (region, name)
    VALUES
        (1, 'Мосавтодор');
    INSERT INTO public.region
        (region, name)
    VALUES
        (2, 'Чукотка');
