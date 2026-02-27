package domain

import "errors"

// Ошибки авторизации
var (
	ErrInvalidToken       = errors.New("невалидный токен")
	ErrInvalidCredentials = errors.New("неверный email или пароль")
)

// Ошибки пользователей
var (
	ErrEmailTaken   = errors.New("email уже занят")
	ErrUserNotFound = errors.New("пользователь не найден")
)

// Ошибки команд
var (
	ErrNoPermission  = errors.New("недостаточно прав")
	ErrAlreadyMember = errors.New("пользователь уже в команде")
	ErrNotTeamMember = errors.New("вы не являетесь членом команды")
	ErrTeamNotFound  = errors.New("команда не найдена")
)

// Ошибки задач
var (
	ErrTaskNotFound = errors.New("задача не найдена")
)
