-- Table: public."statistics"

-- DROP TABLE public."statistics";
DROP table if exists public."statistics";
CREATE TABLE public."statistics"
(
    region integer NOT NULL,
    area integer NOT NULL,
    id integer NOT NULL,
    date date NOT NULL,
    stat jsonb NOT NULL
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;
COMMENT ON TABLE public.statistics
    IS 'Статистика';
COMMENT ON COLUMN public.statistics.region
    IS 'Регион';
COMMENT ON COLUMN public.statistics.area
    IS 'Район';
COMMENT ON COLUMN public.statistics.id
    IS 'Номер перекрестка';
COMMENT ON COLUMN public.statistics.date
    IS 'Дата';
COMMENT ON COLUMN public.statistics.stat
    IS 'Статистика за день';

ALTER TABLE public."statistics"
    OWNER to postgres;