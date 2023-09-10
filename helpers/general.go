package helpers

import "strings"

func RemoveDuplicateStrings(inputSlice []string) []string {
    uniqueElements := make(map[string]bool)
    resultSlice := []string{}

    for _, element := range inputSlice {
        if !uniqueElements[element] {
            uniqueElements[element] = true
            resultSlice = append(resultSlice, element)
        }
    }

    return resultSlice
}

func EscapeString(str string) string {
    quoteEscaper := strings.NewReplacer(`'`, `\'`, `"`, `\"`)
    return quoteEscaper.Replace(str)
}
