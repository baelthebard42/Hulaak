package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) // bytes consists: password, salt and cost info which is retrieved when comparing it with another password
	return string(bytes), err
}

func VerifyPassword(claimedPassword string, realPasswordHash string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(realPasswordHash), []byte(claimedPassword))
	return err == nil
}
