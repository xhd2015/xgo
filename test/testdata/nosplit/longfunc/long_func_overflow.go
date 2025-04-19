package main

import (
	"fmt"
)

// go run -gcflags="all=-N -l" ./test/testdata/nosplit/long_func.go
//
// # command-line-arguments
// main.longFunc: nosplit stack over 792 byte limit
// main.longFunc<1>
//
//	grows 1616 bytes
//	824 bytes over limit
func main() {
	// printFunc(200)
	fmt.Printf("long func nosplit:%v\n", longFunc(1))
}

func printFunc(n int) {
	fmt.Printf("i0:=a\n")
	for i := 1; i < n; i++ {
		fmt.Printf("i%d:=i%d\n", i, i-1)
	}
	fmt.Printf("return i%d\n", n-1)
}

//go:nosplit
func longFunc(a int) int {
	i0 := a
	i1 := i0
	i2 := i1
	i3 := i2
	i4 := i3
	i5 := i4
	i6 := i5
	i7 := i6
	i8 := i7
	i9 := i8
	i10 := i9
	i11 := i10
	i12 := i11
	i13 := i12
	i14 := i13
	i15 := i14
	i16 := i15
	i17 := i16
	i18 := i17
	i19 := i18
	i20 := i19
	i21 := i20
	i22 := i21
	i23 := i22
	i24 := i23
	i25 := i24
	i26 := i25
	i27 := i26
	i28 := i27
	i29 := i28
	i30 := i29
	i31 := i30
	i32 := i31
	i33 := i32
	i34 := i33
	i35 := i34
	i36 := i35
	i37 := i36
	i38 := i37
	i39 := i38
	i40 := i39
	i41 := i40
	i42 := i41
	i43 := i42
	i44 := i43
	i45 := i44
	i46 := i45
	i47 := i46
	i48 := i47
	i49 := i48
	i50 := i49
	i51 := i50
	i52 := i51
	i53 := i52
	i54 := i53
	i55 := i54
	i56 := i55
	i57 := i56
	i58 := i57
	i59 := i58
	i60 := i59
	i61 := i60
	i62 := i61
	i63 := i62
	i64 := i63
	i65 := i64
	i66 := i65
	i67 := i66
	i68 := i67
	i69 := i68
	i70 := i69
	i71 := i70
	i72 := i71
	i73 := i72
	i74 := i73
	i75 := i74
	i76 := i75
	i77 := i76
	i78 := i77
	i79 := i78
	i80 := i79
	i81 := i80
	i82 := i81
	i83 := i82
	i84 := i83
	i85 := i84
	i86 := i85
	i87 := i86
	i88 := i87
	i89 := i88
	i90 := i89
	i91 := i90
	i92 := i91
	i93 := i92
	i94 := i93
	i95 := i94
	i96 := i95
	i97 := i96
	i98 := i97
	i99 := i98
	i100 := i99
	i101 := i100
	i102 := i101
	i103 := i102
	i104 := i103
	i105 := i104
	i106 := i105
	i107 := i106
	i108 := i107
	i109 := i108
	i110 := i109
	i111 := i110
	i112 := i111
	i113 := i112
	i114 := i113
	i115 := i114
	i116 := i115
	i117 := i116
	i118 := i117
	i119 := i118
	i120 := i119
	i121 := i120
	i122 := i121
	i123 := i122
	i124 := i123
	i125 := i124
	i126 := i125
	i127 := i126
	i128 := i127
	i129 := i128
	i130 := i129
	i131 := i130
	i132 := i131
	i133 := i132
	i134 := i133
	i135 := i134
	i136 := i135
	i137 := i136
	i138 := i137
	i139 := i138
	i140 := i139
	i141 := i140
	i142 := i141
	i143 := i142
	i144 := i143
	i145 := i144
	i146 := i145
	i147 := i146
	i148 := i147
	i149 := i148
	i150 := i149
	i151 := i150
	i152 := i151
	i153 := i152
	i154 := i153
	i155 := i154
	i156 := i155
	i157 := i156
	i158 := i157
	i159 := i158
	i160 := i159
	i161 := i160
	i162 := i161
	i163 := i162
	i164 := i163
	i165 := i164
	i166 := i165
	i167 := i166
	i168 := i167
	i169 := i168
	i170 := i169
	i171 := i170
	i172 := i171
	i173 := i172
	i174 := i173
	i175 := i174
	i176 := i175
	i177 := i176
	i178 := i177
	i179 := i178
	i180 := i179
	i181 := i180
	i182 := i181
	i183 := i182
	i184 := i183
	i185 := i184
	i186 := i185
	i187 := i186
	i188 := i187
	i189 := i188
	i190 := i189
	i191 := i190
	i192 := i191
	i193 := i192
	i194 := i193
	i195 := i194
	i196 := i195
	i197 := i196
	i198 := i197
	i199 := i198
	return i199
}
