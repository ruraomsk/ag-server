-- Table: public.history

-- DROP TABLE public.history;
DROP table if exists public.history;

CREATE TABLE public.history
(
    region integer NOT NULL,
    area integer NOT NULL,
    id integer NOT NULL,
    login text COLLATE pg_catalog."default",
    tm timestamp with time zone NOT NULL,
    state jsonb NOT NULL
)

    WITH (
        autovacuum_enabled = TRUE
        )
    TABLESPACE pg_default;

ALTER TABLE public.history
    OWNER to postgres;