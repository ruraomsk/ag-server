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
    state jsonb NOT NULL
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public."cross"
    OWNER to postgres;