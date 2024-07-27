package format

// /*<begin X>*/ ... /*<end X>*/
const BEGIN = "/*<begin "
const END = "/*<end "
const CLOSE = ">*/"

const REPLACED_BEGIN = "/*<replaced>"
const REPLACED_END = "</replaced>*/"

func Begin(id string) string {
	return BEGIN + id + CLOSE
}

func End(id string) string {
	return END + id + CLOSE
}
