#Основной сервер системы (агрегатор)
##Назначение 
    Предназначен для запуска всех остальных компонентов системы и обеспечивает ее функционирование.
Выполняет следующиие действия:
    1. Настраивает систему логирования в текстовый файл (по пути log/ag-server)
    2. Принимает настройки системы из каталога setup/setup.json
    3. Запускает внутреннюю базу данных pudge
    4. Запускает коммункационный сервер
    5. Запускает внутренний аудитор состояния системы

##Внутренняя база pudge
    Предназначена для оперативного хранения состояния контроллеров и сохранения   
