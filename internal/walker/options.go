package walker

import (
	"errors"
	"fmt"
)

// ErrWalkStopped обозначает явную остановку обхода.
// Не используется при обычном break из итератора.
var ErrWalkStopped = errors.New("filesystem walk stopped")

// Options управляет обходом файловой системы.
type Options struct {
	// IncludeFiles включает обычные файлы, ссылки
	// и другие объекты файловой системы.
	// При false выводятся только директории.
	IncludeFiles bool

	// ShowAll отключает все фильтры:
	// скрытые файлы, встроенные исключения,
	// пользовательские шаблоны и .gitignore.
	ShowAll bool

	// MaxDepth ограничивает глубину обхода.
	//
	// -1: без ограничений
	//  0: только корень
	//  1: корень и непосредственные потомки
	MaxDepth int

	// Exclude содержит пользовательские шаблоны.
	// Например: "*.log", "tmp/", "**/cache/**".
	Exclude []string

	// UseGitIgnore включает применение правил
	// из .gitignore, включая вложенные файлы.
	UseGitIgnore bool
}

// DefaultOptions возвращает стандартную конфигурацию.
func DefaultOptions() Options {
	return Options{
		IncludeFiles: true,
		ShowAll:      false,
		MaxDepth:     -1,
		Exclude:      nil,
		UseGitIgnore: false,
	}
}

// Validate проверяет корректность параметров обхода.
func (o Options) Validate() error {
	if o.MaxDepth < -1 {
		return fmt.Errorf(
			"invalid max depth %d: must be >= -1",
			o.MaxDepth,
		)
	}

	for i, pattern := range o.Exclude {
		if pattern == "" {
			return fmt.Errorf(
				"exclude pattern at index %d is empty",
				i,
			)
		}
	}

	return nil
}
