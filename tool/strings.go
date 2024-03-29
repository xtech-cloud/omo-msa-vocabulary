package tool

import (
	"github.com/satori/go.uuid"
)

func SubString(source string, start int, end int) string {
	var r = []rune(source)
	length := len(r)

	if start < 0 || end > length || start > end {
		return ""
	}

	if start == 0 && end == length {
		return source
	}

	return string(r[start:end])
}

func CreateUUID() string {
	u2 := uuid.NewV4()
	// Parsing UUID from string input
	/*u3, err1 := uuid.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err1 != nil {
		fmt.Printf("Something went wrong: %s", err1)
		return ""
	}
	fmt.Printf("Successfully parsed: %s", u3)*/
	return u2.String()
}

func DifferenceStrings(source, dest []string) (string, string) {
	from := ""
	to := ""
	if len(source) == 1 {
		from = source[0]
	} else {
		for _, s := range source {
			if !HasItem(dest, s) {
				from = s
				break
			}
		}
	}
	if len(dest) == 1 {
		to = dest[0]
	} else {
		for _, s := range dest {
			if !HasItem(source, s) {
				to = s
				break
			}
		}
	}
	return from, to
}
