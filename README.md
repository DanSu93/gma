# gma
In memory key-value store

CMD:
* gma serve - запуск сервера, который откроет сокет и будет основным хранилищем key-value store
* gma cli - консоль, в которой можно писать команды, валидные команды будут отправляться в сокет. Если сервер не запущен, то выходим с сообщением что сервер не запущен. 


TODO:
* Start server demonize
* Support TTL
* Tests
* Support persistence storage
* Нагрузочное тестирование
