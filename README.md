# gotree

**gotree** - небольшая CLI-утилита на Go для потокового обхода файловой системы и отображения структуры директорий.

Утилита поддерживает терминальное дерево, машиночитаемые форматы, фильтрацию файлов, правила `.gitignore`, ограничение глубины и сбор статистики.

## Возможности

- Потоковая передача элементов.
- Четыре формата вывода: `tree`, `jsonl`, `json`, `paths`.
- Автоматический выбор формата.
- Отображение только директорий.
- Скрытие служебных файлов и директорий по умолчанию.
- Пользовательские исключения через шаблоны.
- Поддержка вложенных `.gitignore`.
- Ограничение глубины обхода.
- Подсчёт файлов, директорий, символических ссылок и других объектов.
- Передача родительского пути, глубины и типа каждого объекта.

## Требования

- Go 1.23 или новее.
- Поддерживаемая операционная система: Linux, macOS или Windows.

## Установка

Клонируйте репозиторий:

```bash
git clone https://github.com/eternalpaininside/gotree.git
cd gotree
```

Установите зависимости:

```bash
go mod tidy
```

Соберите утилиту:

```bash
go build -o gotree ./cmd/gotree
```

Запустите:

```bash
./gotree -f tree .
```

Также возможен запуск без сборки:

```bash
go run ./cmd/gotree -f tree .
```

## Использование

```bash
gotree [flags] [path]
```

Если путь не указан, используется текущая директория.

### Флаги

| Короткий | Длинный | Описание | По умолчанию |
|---|---|---|---|
| `-a` | `--all` | Отключить все фильтры исключений | `false` |
| `-d` | `--dirs-only` | Показывать только директории | `false` |
| `-f` | `--format` | Формат: `auto`, `tree`, `jsonl`, `json`, `paths` | `auto` |
| `-e` | `--exclude` | Добавить шаблон исключения | Пусто |
| `-L` | `--max-depth` | Максимальная глубина обхода | `-1` |
| `-g` | `--gitignore` | Применить правила `.gitignore` | `false` |
| `-s` | `--stats` | Добавить статистику в stdout | `false` |
| `-h` | `--help` | Показать справку | - |

Флаг `-e` можно передавать многократно.

## Примеры

### Вывод дерева

```bash
gotree -f tree .
```

Пример:

```text
gotree/
├── README.md
├── cmd/
│   └── gotree/
├── go.mod
├── go.sum
└── internal/
    ├── output/
    └── walker/
```

### Только директории

```bash
gotree -d -f tree .
```

### Все файлы, включая скрытые

```bash
gotree -a -f tree .
```

Флаг `-a` имеет приоритет над встроенными исключениями, `-e` и `.gitignore`.

Он не отменяет `-d` и `-L`.

### Ограничение глубины

```bash
gotree -L 2 -f tree .
```

Глубина отсчитывается от корня:

- `0` - только корень;
- `1` - корень и непосредственные потомки;
- `2` - два уровня вложенности;
- `-1` - без ограничения.

### Пользовательские исключения

```bash
gotree -e "*.log" -e "tmp/" -f tree .
```

Поддерживаемые шаблоны:

| Шаблон | Описание |
|---|---|
| `*.log` | Имена с расширением `.log` |
| `tmp/` | Директории `tmp` |
| `internal/cache` | Путь относительно корня обхода |
| `**/cache` | Объект `cache` на любой глубине |
| `**/cache/**` | Содержимое `cache` на произвольной глубине |

Шаблоны с `/` интерпретируются относительно корня обхода. Шаблоны без `/` сопоставляются с именем объекта.

### Поддержка `.gitignore`

```bash
gotree -g -f tree .
```

Поддерживаются:

- Обычные шаблоны Git.
- Шаблоны с `*` и `**`.
- Отрицания через `!`.
- Вложенные `.gitignore`.
- Области действия правил относительно соответствующей директории.

По умолчанию `.gitignore` не применяется.

При одновременном использовании `-a -g` фильтрация отключается полностью.

### Статистика

```bash
gotree -s -f tree .
```

Пример:

```text
gotree/
├── README.md
├── go.mod
└── internal/
    └── walker/
        └── walker.go

Files: 3, Directories: 3, Symlinks: 0, Other: 0, Total: 6
```

При `-d` счётчик файлов равен нулю.

## Форматы вывода

### tree

Человекочитаемое дерево с символами `├──`, `└──` и `│`.

```bash
gotree -f tree .
```

### jsonl

JSON Lines - один JSON-объект на строку.

```bash
gotree -f jsonl .
```

Пример:

```jsonl
{"path":"/project","name":"project","depth":0,"type":"directory","is_last":true}
{"path":"/project/main.go","name":"main.go","parent":"/project","depth":1,"type":"file","is_last":true}
```

При `-s` добавляется отдельная запись со статистикой, содержащая поле `record_type: "stats"`.

Пример обработки через `jq`:

```bash
gotree -f jsonl . | jq 'select(.type == "file")'
```

### json

Единый JSON-документ с массивом элементов.

```bash
gotree -f json .
```

С добавлением статистики:

```bash
gotree -f json -s . | jq .
```

Пример структуры:

```json
{
  "entries": [
    {
      "path": "/project",
      "name": "project",
      "depth": 0,
      "type": "directory",
      "is_last": true
    }
  ],
  "stats": {
    "files": 0,
    "directories": 1,
    "symlinks": 0,
    "other": 0
  }
}
```

### paths

Относительные пути, по одному на строку.

```bash
gotree -f paths .
```

Пример:

```text
.
cmd
cmd/gotree
cmd/gotree/main.go
internal
internal/walker
internal/walker/walker.go
```

Использование с `grep`:

```bash
gotree -f paths . | grep '\.go$'
```

При `-s` добавляется строка статистики с префиксом `#`.

### Автоматический формат

По умолчанию используется `auto`:

- Если `stdout` подключён к терминалу, выбирается `tree`.
- Если вывод перенаправлен в файл или пайплайн, выбирается `jsonl`.

Например:

```bash
gotree . | jq .
```

## Фильтрация по умолчанию

Скрываются файлы и директории, имена которых начинаются с точки:

```text
.git/
.idea/
.DS_Store
.env
.gitignore
```

Дополнительно исключаются:

```text
node_modules/
__pycache__/
venv/
dist/
build/
coverage/
```

## Архитектура

```text
gotree/
├── cmd/
│   └── gotree/
│       ├── main.go
│       └── main_test.go
├── internal/
│   ├── output/
│   │   ├── json.go
│   │   └── output_test.go
│   │   ├── paths.go
│   │   ├── pipeline.go
│   │   ├── terminal.go
│   ├── stats/
│   │   ├── stats.go
│   │   └── stats_test.go
│   └── walker/
│       ├── filter/
│       │   ├── filter.go
│       │   ├── filter_test.go
│       │   ├── gitignore.go
│       │   └── gitignore_test.go
│       ├── options.go
│       ├── walker.go
│       └── walker_test.go
├── go.mod
├── go.sum
└── README.md
```