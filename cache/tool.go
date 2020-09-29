package cache

func hadItem(array []string, target string) bool {
	if array == nil || len(array) < 1 {
		return false
	}
	for _, s := range array {
		if s == target {
			return true
		}
	}
	return false
}