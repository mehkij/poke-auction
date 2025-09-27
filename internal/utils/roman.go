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

func ToNum(numeral string) int {
	switch numeral {
	case "i":
		return 1
	case "ii":
		return 2
	case "iii":
		return 3
	case "iv":
		return 4
	case "v":
		return 5
	case "vi":
		return 6
	case "vii":
		return 7
	case "viii":
		return 8
	case "ix":
		return 9
	default:
		return 0
	}
}
