-- Table: public.routes

DROP TABLE if exists public.routes;
DROP SEQUENCE if exists id_routes;
CREATE TABLE public.routes
(
	region integer NOT NULL,
    description text COLLATE pg_catalog."default",
    box jsonb,
    listtl jsonb
)

WITH (
    autovacuum_enabled = TRUE
)
TABLESPACE pg_default;
COMMENT ON TABLE public."routes"
    IS 'Описание маршрутов';
COMMENT ON COLUMN public.routes.region
    IS 'Регион';
COMMENT ON COLUMN public.routes.description
    IS 'Описание';
COMMENT ON COLUMN public.routes.box
    IS 'Координы маршрута';
COMMENT ON COLUMN public.routes.listtl
    IS 'Перечень входящих перекрестков';


ALTER TABLE public.routes
    OWNER to postgres;
insert into public.routes ( region,  description, box, listtl) VALUES (1,'123312','{"point0": {"X": 36.01732914160939, "Y": 55.503656600413144}, "point1": {"X": 36.05902950415313, "Y": 55.506676662403414}}','[{"num": 0, "pos": {"id": 51, "area": "1", "region": "1"}, "phase": 1, "point": {"X": 36.01732914160939, "Y": 55.506676662403414}}, {"num": 1, "pos": {"id": 50, "area": "1", "region": "1"}, "phase": 1, "point": {"X": 36.02021146926837, "Y": 55.50629687806578}}, {"num": 2, "pos": {"id": 49, "area": "1", "region": "1"}, "phase": 1, "point": {"X": 36.02911103877978, "Y": 55.50496487779033}}, {"num": 3, "pos": {"id": 48, "area": "1", "region": "1"}, "phase": 1, "point": {"X": 36.03557076843429, "Y": 55.50387490015615}}, {"num": 4, "pos": {"id": 47, "area": "1", "region": "1"}, "phase": 1, "point": {"X": 36.04531720276914, "Y": 55.503656600413144}}, {"num": 5, "pos": {"id": 46, "area": "1", "region": "1"}, "phase": 1, "point": {"X": 36.05902950415313, "Y": 55.50368892386935}}]');
insert into public.routes ( region,  description, box, listtl) VALUES (1,'fsr23','{"point0": {"X": 37.90669784183896, "Y": 55.97158806009099}, "point1": {"X": 37.93845701678179, "Y": 55.976127661984606}}','[{"num": 0, "pos": {"id": 95, "area": "2", "region": "1"}, "phase": 1, "point": {"X": 37.93845701678179, "Y": 55.97158806009099}}, {"num": 1, "pos": {"id": 93, "area": "2", "region": "1"}, "phase": 1, "point": {"X": 37.92425226357476, "Y": 55.97170171348457}}, {"num": 2, "pos": {"id": 91, "area": "2", "region": "1"}, "phase": 1, "point": {"X": 37.90669784183896, "Y": 55.976127661984606}}]');
insert into public.routes ( region,  description, box, listtl) VALUES (1,'fsr232','{"point0": {"X": 37.86810720595662, "Y": 55.97158806009099}, "point1": {"X": 37.93845701678179, "Y": 55.99554449626898}}','[{"num": 0, "pos": {"id": 95, "area": "2", "region": "1"}, "phase": 1, "point": {"X": 37.93845701678179, "Y": 55.97158806009099}}, {"num": 1, "pos": {"id": 93, "area": "2", "region": "1"}, "phase": 1, "point": {"X": 37.92425226357476, "Y": 55.97170171348457}}, {"num": 2, "pos": {"id": 91, "area": "2", "region": "1"}, "phase": 1, "point": {"X": 37.90669784183896, "Y": 55.976127661984606}}, {"num": 3, "pos": {"id": 3, "area": "2", "region": "1"}, "phase": 1, "point": {"X": 37.86810720595662, "Y": 55.99554449626898}}, {"num": 4, "pos": {"id": 2, "area": "2", "region": "1"}, "phase": 1, "point": {"X": 37.86819230289014, "Y": 55.98728647405042}}]');
insert into public.routes ( region,  description, box, listtl) VALUES (1,'ЗУ 1 (ул.Московская-ш.Можайское)','{"point0": {"X": 36.01732914160939, "Y": 55.503656600413144}, "point1": {"X": 36.05902950415313, "Y": 55.506676662403414}}','[{"num": 0, "pos": {"id": 51, "area": "1", "region": "1"}, "phase": 1, "point": {"X": 36.01732914160939, "Y": 55.506676662403414}}, {"num": 1, "pos": {"id": 50, "area": "1", "region": "1"}, "phase": 1, "point": {"X": 36.02021146926837, "Y": 55.50629687806578}}, {"num": 2, "pos": {"id": 49, "area": "1", "region": "1"}, "phase": 1, "point": {"X": 36.02911103877978, "Y": 55.50496487779033}}, {"num": 3, "pos": {"id": 48, "area": "1", "region": "1"}, "phase": 1, "point": {"X": 36.03557076843429, "Y": 55.50387490015615}}, {"num": 4, "pos": {"id": 47, "area": "1", "region": "1"}, "phase": 1, "point": {"X": 36.04531720276914, "Y": 55.503656600413144}}, {"num": 5, "pos": {"id": 46, "area": "1", "region": "1"}, "phase": 1, "point": {"X": 36.05902950415313, "Y": 55.50368892386935}}]');



