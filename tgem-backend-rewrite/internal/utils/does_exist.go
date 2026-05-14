package utils

func DoesExist(list []string, entry string) bool {
  for _, value := range list {
    if value == entry {
      return true
    }
  }

  return false
}
