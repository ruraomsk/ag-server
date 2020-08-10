-- Table: public.status

-- DROP TABLE public.status;
DROP table if exists public.status; 
CREATE TABLE public.status
(
    id integer NOT NULL,
    description text NOT NULL
)
WITH (
    OIDS = FALSE,
    autovacuum_enabled = TRUE
)
TABLESPACE pg_default;

ALTER TABLE public.status
    OWNER to postgres;
COMMENT ON TABLE public.status
    IS 'Описание статуса устройств';

insert into public.status (id,description) values (1,'Координированное управление');
insert into public.status (id,description) values (2,'Диспетчерское управление');
insert into public.status (id,description) values (3,'Ручное управление');
insert into public.status (id,description) values (4,'Перекрёсток находится на маршруте «Зелёной улицы»');
insert into public.status (id,description) values (5,'Локальное управление');
insert into public.status (id,description) values (6,'КУ ЖМ - желтое мигание по расписанию');
insert into public.status (id,description) values (7,'ДУ ЖМ - желтое мигание заданное из центра(АРМа)');
insert into public.status (id,description) values (8,'РУ ЖМ - желтое мигание заданное на перекрестке');
insert into public.status (id,description) values (9,'ЛУ ЖМ - желтое мигание по расписанию ДК');
insert into public.status (id,description) values (10,'КУ КК - кругом Красный');
insert into public.status (id,description) values (11,'КУ ОС - отключение светофора по расписанию');
insert into public.status (id,description) values (12,'ДУ ОС - отключение светофора заданное из цетра(АРМа)');
insert into public.status (id,description) values (13,'РУ ОС - отключение светофора заданное на перекрестке');
insert into public.status (id,description) values (14,'ЛУ ОС - отключение светофора по расписанию ДК');
insert into public.status (id,description) values (15,'Открытые двери на ДК');
insert into public.status (id,description) values (16,'Авария 220В');
insert into public.status (id,description) values (17,'Выключен УСДК/ДК');
insert into public.status (id,description) values (18,'Нет связи с УСДК/ДК');
insert into public.status (id,description) values (19,'Нет связи с ПСПД');
insert into public.status (id,description) values (20,'Обрыв ЛС КЗЦ');
insert into public.status (id,description) values (21,'Превышение трафика');
insert into public.status (id,description) values (22,'Базовая привязка');
insert into public.status (id,description) values (23,'Неисправность часов или GPS');
insert into public.status (id,description) values (24,'Коррекция привязки');
insert into public.status (id,description) values (25,'Несуществующая фаза');
insert into public.status (id,description) values (26,'Несуществующий код');
insert into public.status (id,description) values (27,'Координированное управление и перегоревшая лампа');
insert into public.status (id,description) values (28,'Обрыв линий связи');
insert into public.status (id,description) values (29,'Негоден по паритету (ошибка контрольной суммы)');
insert into public.status (id,description) values (30,'Отключен светофор из-за конфликта направлений');
insert into public.status (id,description) values (31,'Конфликт направлений');
insert into public.status (id,description) values (32,'Желтое мигание из-за перегорания контролируемых красных ламп');
insert into public.status (id,description) values (33,'Негоден по перегоранию контр. Ламп');
insert into public.status (id,description) values (34,'Не включается в координацию');
insert into public.status (id,description) values (35,'Дорожный контроллер не подчиняется командам');
insert into public.status (id,description) values (36,'Длинный промежуточный такт');
insert into public.status (id,description) values (37,'Обрыв линий связи ЭВМ с перекрестками');
insert into public.status (id,description) values (38,'Нет информации о работе перекрестка');
insert into public.status (id,description) values (39,'Нет данных о работе перекрестка');
