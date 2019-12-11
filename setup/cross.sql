-- Table: public."cross"

-- DROP TABLE public."cross";
DROP table if exists public.cross;
CREATE TABLE public.cross
(
    region integer NOT NULL,
    id integer NOT NULL,
    idevice int NOT NULL,
    device jsonb NOT NULL

)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.cross
    OWNER to postgres;