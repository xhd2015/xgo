package coverage

// Merge profiles
// the result will be compacted
func Merge(covs ...[]*CovLine) []*CovLine {
	if len(covs) == 0 {
		return nil
	}
	if len(covs) == 1 {
		return Compact(covs[0])
	}
	result := covs[0]
	lineByPrefix := make(map[string]*CovLine, len(result))
	result = compact(lineByPrefix, result)
	for i := 1; i < len(covs); i++ {
		result = doMerge(lineByPrefix, result, covs[i])
	}
	return result
}

func Filter(covs []*CovLine, check func(line *CovLine) bool) []*CovLine {
	n := len(covs)
	j := 0
	for i := 0; i < n; i++ {
		if check(covs[i]) {
			covs[j] = covs[i]
			j++
		}
	}
	return covs[:j]
}

func doMerge(lineByPrefix map[string]*CovLine, dst []*CovLine, src []*CovLine) []*CovLine {
	for _, line := range src {
		prevLine, ok := lineByPrefix[line.Prefix]
		if ok {
			prevLine.Count += line.Count
			continue
		}
		lineByPrefix[line.Prefix] = line
		dst = append(dst, line)
	}
	return dst
}

func findByPrefix(linesA []*CovLine, prefix string) int {
	for i := 0; i < len(linesA); i++ {
		if linesA[i].Prefix == prefix {
			return i
		}
	}
	return -1
}
