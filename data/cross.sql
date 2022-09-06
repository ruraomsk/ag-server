-- Table: public."cross"

-- DROP TABLE public."cross";
DROP table if exists public."cross";
CREATE TABLE public."cross"
(
    region integer NOT NULL,
    area integer NOT NULL,
    subarea integer NOT NULL,
    id integer NOT NULL,
    idevice int NOT NULL,
    dgis point NOT NULL,
    describ text,
    status int NOT NULL,
    state jsonb NOT NULL,
    link jsonb
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;
COMMENT ON TABLE public."cross"
    IS 'Описание перекрестков';
COMMENT ON COLUMN public.cross.region
    IS 'Регион';
COMMENT ON COLUMN public.cross.area
    IS 'Район';
COMMENT ON COLUMN public.cross.subarea
    IS 'Подрайон';
COMMENT ON COLUMN public.cross.id
    IS 'Номер перекрестка';
COMMENT ON COLUMN public.cross.idevice
    IS 'Контроллер';
COMMENT ON COLUMN public.cross.dgis
    IS 'Координаты';
COMMENT ON COLUMN public.cross.describ
    IS 'Описание';
COMMENT ON COLUMN public.cross.status
    IS 'Состояние';
COMMENT ON COLUMN public.cross.state
    IS 'Информация';


ALTER TABLE public."cross"
    OWNER to postgres;