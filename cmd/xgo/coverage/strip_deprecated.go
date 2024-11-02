package coverage

import (
	"io"
	"strings"
)

// this stripWriter is historically used to convert some string to another
// now it is not necessary, left here only for reference purpose

type stripWriter struct {
	buf []byte
	w   io.Writer
}

func (c *stripWriter) Write(p []byte) (int, error) {
	n := len(p)
	for i := 0; i < n; i++ {
		if p[i] != '\n' {
			c.buf = append(c.buf, p[i])
			continue
		}
		err := c.sendLine()
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (c *stripWriter) sendLine() error {
	str := string(c.buf)
	c.buf = nil
	for _, stripPair := range stripPairs {
		from := string(stripPair[0])
		to := string(stripPair[1])
		str = strings.ReplaceAll(str, from, to)
	}

	_, err := c.w.Write([]byte(str))
	if err != nil {
		return err
	}
	_, err = c.w.Write([]byte{'\n'})
	return err
}

var stripPairs = [][2][]byte{
	{
		[]byte{103},
		[]byte{115},
	},
}

func (c *stripWriter) Close() error {
	if len(c.buf) > 0 {
		c.sendLine()
	}
	return nil
}
