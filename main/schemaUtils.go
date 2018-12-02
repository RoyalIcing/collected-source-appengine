package main

func optionalStrings(items []string) []*string {
	out := make([]*string, 0, len(items))
	for _, item := range items {
		localItem := item
		out = append(out, &localItem)
	}
	return out
}
