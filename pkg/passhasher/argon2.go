package passhasher

import "github.com/alexedwards/argon2id"

var passwordParams = &argon2id.Params{
	Memory:      19 * 1024,
	Iterations:  2,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

type Argon2Hasher struct{}

func (Argon2Hasher) Hash(password string) (string, error) {
	return argon2id.CreateHash(password, passwordParams)
}

func (Argon2Hasher) Compare(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
