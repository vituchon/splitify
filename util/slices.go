package util

func Filter[T any](items []T, fn func(item T) bool) []T {
    filteredItems := []T{}
    for _, value := range items {
        if fn(value) {
            filteredItems = append(filteredItems, value)
        }
    }
    return filteredItems
}


func ToValues[T any](items []*T) []T {
    values := []T{}
    for _, value := range items {
        values = append(values, *value)
    }
    return values
}

