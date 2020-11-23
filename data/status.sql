-- Table: public.status

-- DROP TABLE public.status;
DROP table if exists public.status; 
CREATE TABLE public.status
(
    id integer NOT NULL,
    description text NOT NULL,
    control boolean not null
)
WITH (
    OIDS = FALSE,
    autovacuum_enabled = TRUE
)
TABLESPACE pg_default;

ALTER TABLE public.status
    OWNER to postgres;
COMMENT ON TABLE public.status
    IS 'Справочник статусов устройств';
COMMENT ON COLUMN public.status.id
    IS 'Код состоянния';
COMMENT ON COLUMN public.status.description
    IS 'Описание состоянния';
COMMENT ON COLUMN public.status.control
    IS 'Признак возможности управления';

insert into public.status (id,description,control) values (1,'Координированное управление',TRUE);
insert into public.status (id,description,control) values (2,'Диспетчерское управление',TRUE);
insert into public.status (id,description,control) values (3,'Ручное управление',FALSE);
insert into public.status (id,description,control) values (4,'Перекрёсток находится на маршруте «Зелёной улицы»',TRUE);
insert into public.status (id,description,control) values (5,'Локальное управление',TRUE);
insert into public.status (id,description,control) values (6,'КУ ЖМ - желтое мигание по расписанию',TRUE);
insert into public.status (id,description,control) values (7,'ДУ ЖМ - желтое мигание заданное из центра(АРМа)',TRUE);
insert into public.status (id,description,control) values (8,'РУ ЖМ - желтое мигание заданное на перекрестке',FALSE);
insert into public.status (id,description,control) values (9,'ЛУ ЖМ - желтое мигание по расписанию ДК',TRUE);
insert into public.status (id,description,control) values (10,'КУ КК - кругом Красный',TRUE);
insert into public.status (id,description,control) values (11,'КУ ОС - отключение светофора по расписанию',TRUE);
insert into public.status (id,description,control) values (12,'ДУ ОС - отключение светофора заданное из цетра(АРМа)',TRUE);
insert into public.status (id,description,control) values (13,'РУ ОС - отключение светофора заданное на перекрестке',FALSE);
insert into public.status (id,description,control) values (14,'ЛУ ОС - отключение светофора по расписанию ДК',TRUE);
insert into public.status (id,description,control) values (15,'Открытые двери на ДК',TRUE);
insert into public.status (id,description,control) values (16,'Авария 220В',FALSE);
insert into public.status (id,description,control) values (17,'Выключен УСДК/ДК',FALSE);
insert into public.status (id,description,control) values (18,'Нет связи с УСДК/ДК',FALSE);
insert into public.status (id,description,control) values (19,'Нет связи с ПСПД',FALSE);
insert into public.status (id,description,control) values (20,'Обрыв ЛС КЗЦ',FALSE);
insert into public.status (id,description,control) values (21,'Превышение трафика',TRUE);
insert into public.status (id,description,control) values (22,'Базовая привязка',TRUE);
insert into public.status (id,description,control) values (23,'Неисправность часов или GPS',TRUE);
insert into public.status (id,description,control) values (24,'Коррекция привязки',FALSE);
insert into public.status (id,description,control) values (25,'Несуществующая фаза',TRUE);
insert into public.status (id,description,control) values (26,'Несуществующий код',TRUE);
insert into public.status (id,description,control) values (27,'Координированное управление и перегоревшая лампа',TRUE);
insert into public.status (id,description,control) values (28,'Обрыв линий связи',FALSE);
insert into public.status (id,description,control) values (29,'Негоден по паритету (ошибка контрольной суммы)',TRUE);
insert into public.status (id,description,control) values (30,'Отключен светофор из-за конфликта направлений',FALSE);
insert into public.status (id,description,control) values (31,'Конфликт направлений',FALSE);
insert into public.status (id,description,control) values (32,'Желтое мигание из-за перегорания контролируемых красных ламп',FALSE);
insert into public.status (id,description,control) values (33,'Негоден по перегоранию контр. Ламп',FALSE);
insert into public.status (id,description,control) values (34,'Не включается в координацию',TRUE);
insert into public.status (id,description,control) values (35,'Дорожный контроллер не подчиняется командам',TRUE);
insert into public.status (id,description,control) values (36,'Длинный промежуточный такт',TRUE);
insert into public.status (id,description,control) values (37,'Обрыв линий связи ЭВМ с перекрестками',FALSE);
insert into public.status (id,description,control) values (38,'Нет информации о работе перекрестка',FALSE);
insert into public.status (id,description,control) values (39,'Нет данных о работе перекрестка',FALSE);
