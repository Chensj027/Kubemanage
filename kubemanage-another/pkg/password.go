package pkg

import (
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost 密码哈希强度。12 在安全与登录耗时(约 200-300ms)之间取平衡。
// 注意：bcrypt 把 cost 编码进哈希本身，改此值只影响“新设置/重置”的密码；
// 已有用户的哈希仍是旧 cost，需改一次密码才会应用新值（见登录耗时优化说明）。
const bcryptCost = 12

// GenSaltPassword 将指定字符串，进行bcrypt编码，用于密码加密
func GenSaltPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

// CheckPassword 检查 用户输入的密码 与 db的hashPassword，是否一致
func CheckPassword(password, hashPassword string) bool {
	// 密码是bcrypt加密的，所以要先对password加密，然后再与db比对
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}

// PasswordNeedsUpgrade 判断已存哈希的 cost 是否与当前设定不一致，需要重算。
// 用于登录成功后透明升级旧 cost 的密码哈希。
func PasswordNeedsUpgrade(hashPassword string) bool {
	cost, err := bcrypt.Cost([]byte(hashPassword))
	if err != nil {
		return false
	}
	return cost != bcryptCost
}
