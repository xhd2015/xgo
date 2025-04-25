//go:build ignore

package work

import (
	"encoding/json"
	"os"
	"sync"
)

var xgoPkgTrapSum map[string]string
var xgoSumInitOnce sync.Once

func getXgoPackageTrapSum(pkgPath string) string {
	return getXgoTrapSum()[pkgPath]
}

func getXgoTrapSum() map[string]string {
	xgoSumInitOnce.Do(func() {
		sumFile := os.Getenv("XGO_COMPILER_SYNTAX_REWRITE_PACKAGES_SUM_FILE")
		if sumFile != "" {
			// ignore any err
			sumData, _ := os.ReadFile(sumFile)
			_ = json.Unmarshal(sumData, &xgoPkgTrapSum)
		}
	})
	return xgoPkgTrapSum
}
