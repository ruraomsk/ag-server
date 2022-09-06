-- Table: public.loghistory

-- DROP TABLE public.loghistory;
DROP table if exists public.loghistory;

CREATE TABLE public.loghistory
(
    tm timestamp
    with time zone NOT NULL,
    id integer NOT NULL,
    crossinfo jsonb NOT NULL,
    txt text COLLATE pg_catalog."default" NOT NULL,
    journal jsonb not null,

    region integer NOT NULL
)

    WITH
    (
        autovacuum_enabled = TRUE
        )
    TABLESPACE pg_default;

    ALTER TABLE public.loghistory
    OWNER to postgres;