package util

func CheckIfHeadquater(swiftCode string) bool {
	return swiftCode[len(swiftCode)-3:] == `XXX`
}

func GetPrefix(swiftCode string) string {
	return swiftCode[:len(swiftCode)-3]
}
