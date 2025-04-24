package utils

func ToRoman(number int) string {
	switch number {
	case 1:
		return "i"
	case 2:
		return "ii"
	case 3:
		return "iii"
	case 4:
		return "iv"
	case 5:
		return "v"
	case 6:
		return "vi"
	case 7:
		return "vii"
	case 8:
		return "viii"
	default:
		return ""
	}
}
