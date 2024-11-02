package git

// check if head is ref's ancestor,  can be found following head's first parent
// just like git merge-base --is-ancestor ref head, but excluding merge points
//
// see help for 'git rev-list': https://git-scm.com/docs/git-rev-list
// algorithm:
//
//	find first parent set of head, and exclude first parent set of ref^1
//	    S=git rev-list --first-parent head ^ref^1
//	if ref in S, then head is a direct parent of ref
//
// the direction is ref -> head
func IsFirstParentAncestorOf(dir string, head string, ref string) (bool, error) {
	ok, _, err := FindFirstParentCommitsInBetween(dir, ref, head)
	return ok, err
}

func IsFirstParentReachable(dir string, from string, to string) (bool, error) {
	ok, _, err := FindFirstParentCommitsInBetween(dir, from, to)
	return ok, err
}
