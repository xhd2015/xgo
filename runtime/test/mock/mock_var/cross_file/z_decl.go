package cross_file

import (
	"sync"
	"time"
)

var timeout = 2 * time.Second

var mu sync.Mutex

var count int
