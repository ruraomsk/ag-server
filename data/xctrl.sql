-- Table: public."xctrl"

-- DROP TABLE public."xctrl";
DROP table if exists public."xctrl";
CREATE TABLE public."xctrl"
(
    region integer NOT NULL,
    area integer NOT NULL,
    subarea integer NOT NULL,
    state jsonb not null
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

COMMENT ON TABLE public."xctrl"
    IS 'Описание перекрестков';
COMMENT ON COLUMN public.xctrl.region
    IS 'Регион';
COMMENT ON COLUMN public.xctrl.area
    IS 'Район';
COMMENT ON COLUMN public.xctrl.subarea
    IS 'Подрайон';
COMMENT ON COLUMN public.xctrl.state
    IS 'Прогрмма переключения';

ALTER TABLE public."xctrl"
    OWNER to postgres;

insert into public.xctrl ( region, area, subarea, state) VALUES (1,1,1,'{"rem": 10, "area": 1, "step": 10,  "ltime": "2020-08-07T13:27:27.2246101+06:00",
"pknow": 0, "pklast": 0, "region": 1, "subarea": 1, "left":0.8,"right":1.2,"status": [], "Results":[],"switch": true, "release": true,
"Strategys": [{"pkl": 1,"pks": 2,"pkr": 3, "xleft": 0, "xright": 100},
              {"pkl": 4,"pks": 5,"pkr": 6, "xleft": 100, "xright": 800},
              {"pkl": 7,"pks": 8,"pkr": 9, "xleft": 800, "xright": 9999}],
"Calculates": [ {"region": 1,"id": 8, "area": 1, "chanL": [1,2,3],"chanR": [4] },
                {"region": 1,"id": 7, "area": 1, "chanL": [1,2,3],"chanR": [4,5] },
                {"region": 1,"id": 6, "area": 1, "chanL": [1,2,3],"chanR": [4,5] },
                {"region": 1,"id": 5, "area": 1, "chanL": [1,2,3],"chanR": [4,5] }
]}');
insert into public.xctrl ( region, area, subarea, state) VALUES (1,1,3,'{"rem": 10, "area": 1, "step": 10, "xnum": 205, "ltime": "2020-08-07T13:27:27.2246101+06:00",
  "pknow": 0, "pklast": 0, "region": 1, "subarea": 3, "left":0.8,"right":1.2,"status": [], "Results":[],"switch": true, "release": true,
  "Strategys": [{"pkl": 9,"pks": 8,"pkr": 7, "xleft": 0, "xright": 100},
    {"pkl": 6,"pks": 5,"pkr": 4, "xleft": 100, "xright": 800},
    {"pkl": 3,"pks": 2,"pkr": 1, "xleft": 800, "xright": 9999}],
  "Calculates": [ {"region": 1,"id": 8, "area": 1, "chanL": [1,2,3],"chanR": [4] },
    {"region": 1,"id": 7, "area": 1, "chanL": [1,2,3],"chanR": [4,5] },
    {"region": 1,"id": 6, "area": 1, "chanL": [1,2,3],"chanR": [4,5] },
    {"region": 1,"id": 5, "area": 1, "chanL": [1,2,3],"chanR": [4,5] }
  ]}');

