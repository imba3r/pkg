package bytesize_test

import (
	"math"
	"testing"

	"github.com/imba3r/pkg/assert"
	"github.com/imba3r/pkg/bytesize"
)

func TestByteSize(t *testing.T) {
	assert.Equals(t, "333 B", bytesize.ByteSize(333).String())
	assert.Equals(t, "1.00 KB", bytesize.ByteSize(1024).String())
	assert.Equals(t, "1.00 MB", bytesize.ByteSize(math.Pow(1024, 2)).String())
	assert.Equals(t, "1.00 GB", bytesize.ByteSize(math.Pow(1024, 3)).String())
	assert.Equals(t, "1.00 TB", bytesize.ByteSize(math.Pow(1024, 4)).String())
	assert.Equals(t, "1.00 PB", bytesize.ByteSize(math.Pow(1024, 5)).String())
	assert.Equals(t, "1.00 EB", bytesize.ByteSize(math.Pow(1024, 6)).String())
	assert.Equals(t, "1.00 ZB", bytesize.ByteSize(math.Pow(1024, 7)).String())
	assert.Equals(t, "1.00 YB", bytesize.ByteSize(math.Pow(1024, 8)).String())
}
