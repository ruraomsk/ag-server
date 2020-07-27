-- Table: public.routes

DROP TABLE if exists public.routes;
DROP SEQUENCE if exists id_routes;
CREATE SEQUENCE id_routes START 1;
CREATE TABLE public.routes
(
	region integer NOT NULL,
    id integer NOT NULL DEFAULT nextval('id_routes'),
    description text COLLATE pg_catalog."default",
    box jsonb,
    listtl jsonb
)

WITH (
    autovacuum_enabled = TRUE
)
TABLESPACE pg_default;

ALTER TABLE public.routes
    OWNER to postgres;