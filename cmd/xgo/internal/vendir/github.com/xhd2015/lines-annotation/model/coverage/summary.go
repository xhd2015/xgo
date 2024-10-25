package coverage

type Mode string

const (
	Mode_Line    Mode = "line"
	Mode_Block   Mode = "block"
	Mode_Stmt    Mode = "stmt"
	Mode_Func    Mode = "func"
	Mode_File    Mode = "file"
	Mode_Package Mode = "package"
)

type Summary struct {
	// the default coverage
	*Detail

	Details map[Mode]*Detail `json:"details,omitempty"`
}

type Detail struct {
	// default
	Total       *Item `json:"total,omitempty"`
	Incrimental *Item `json:"incrimental,omitempty"`
}

type Item struct {
	Value         string           `json:"value"`
	ValueNum      float64          `json:"valueNum"`
	Pass          bool             `json:"pass"`
	Total         int              `json:"total"`
	Covered       int              `json:"covered"`
	Threshold     string           `json:"threshold"`
	UncoveredList []*UncoveredItem `json:"uncoveredList"`
}

type UncoveredItem struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Func string `json:"func"`
}
