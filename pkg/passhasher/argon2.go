// Package passhasher предоставляет инструменты для безопасного хеширования
// и верификации паролей с использованием алгоритма Argon2id.
package passhasher

import "github.com/alexedwards/argon2id"

var passwordParams = &argon2id.Params{
	Memory:      19 * 1024, // 19 MB оперативной памяти
	Iterations:  2,         // 2 прохода по памяти
	Parallelism: 1,         // 1 поток выполнения
	SaltLength:  16,        // 16 байт для криптографически стойкой соли
	KeyLength:   32,        // 32 байта для результирующего ключа (хеша)
}

// Argon2Hasher реализует механизмы создания и сверки паролей
// на базе современного криптографического алгоритма Argon2id.
type Argon2Hasher struct{}

// Hash генерирует безопасный Argon2id-хеш из строки пароля на основе
// предопределенных конфигурационных параметров пакета.
// Возвращает готовую строку хеша, включающую в себя соль и параметры алгоритма.
func (Argon2Hasher) Hash(password string) (string, error) {
	return argon2id.CreateHash(password, passwordParams)
}

// Compare выполняет сравнение пароля в открытом виде с предоставленным хешем.
// Сравнение происходит за константное время (constant-time), защищая приложение
// от атак по времени (timing attacks). Возвращает true, если пароли совпадают.
func (Argon2Hasher) Compare(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
