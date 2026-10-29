package domain

import "errors"

// Erros padronizados de domínio para desacoplar regras de negócio da camada HTTP/Banco.
var (
	ErrEmailAlreadyExists        = errors.New("este e-mail já está cadastrado")
	ErrInvalidCredentials        = errors.New("e-mail ou senha inválidos")
	ErrUserNotFound              = errors.New("usuário não encontrado")
	ErrPlanNotFound              = errors.New("plano não encontrado ou inativo")
	ErrActiveSubscriptionExists  = errors.New("usuário já possui uma assinatura ativa")
	ErrSubscriptionNotFound      = errors.New("nenhuma assinatura encontrada para este usuário")
	ErrSubscriptionAlreadyCancel = errors.New("esta assinatura já está cancelada")
)

