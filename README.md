# biathlon_competitions_system

Данный проект представляет собой `Прототип системы для соревнований по биатлону`

Для запуска стоит ввести в консоль команду **(для команд требуется находиться в папке с проектом)**:
```
go run * <config.json> <events.txt>
```

- <config.json> - передайте путь до конфиг файла
- <events.txt> - передайте путь до файла с входящими событиями

## Функционал

Система отвечает сообщением на каждое входящее событие, а также в конце выводит итоговый отчёт по все участникам.

В репозитории присутствуют файлы, на которых можно запустить систему командами **(для команд требуется находиться в папке с проектом)**:
```
go run . config.json events
```

```
go run . config_test.json events_test.txt
```

## Конфигурация (json)

- **Laps**        - Amount of laps for main distance
- **LapLen**      - Length of each main lap
- **PenaltyLen**  - Length of each penalty lap
- **FiringLines** - Number of firing lines per lap
- **Start**       - Planned start time for the first competitor
- **StartDelta**  - Planned interval between starts

## События
All events are characterized by time and event identifier. Outgoing events are events created during program operation. Events related to the "incoming" category cannot be generated and are output in the same form as they were submitted in the input file.

- All events occur sequentially in time. (***Time of event N+1***) >= (***Time of event N***)
- Time format ***[HH:MM:SS.sss]***. Trailing zeros are required in input and output

#### Стандартный формат для события:
[***time***] **eventID** **competitorID** extraParams

