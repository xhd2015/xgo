package coverage

func Merge(covs ...[]*CovLine) []*CovLine {
	if len(covs) == 0 {
		return nil
	}
	if len(covs) == 1 {
		return covs[0]
	}
	result := covs[0]
	for i := 1; i < len(covs); i++ {
		result = doMerge(result, covs[i])
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

func doMerge(linesA []*CovLine, linesB []*CovLine) []*CovLine {
	for _, line := range linesB {
		idx := -1
		for i := 0; i < len(linesA); i++ {
			if linesA[i].Prefix == line.Prefix {
				idx = i
				break
			}
		}
		if idx < 0 {
			linesA = append(linesA, line)
		} else {
			// fmt.Printf("add %s %d %d\n", a[idx].prefix, a[idx].count, line.count)
			linesA[idx].Count += line.Count
		}
	}
	return linesA
}
