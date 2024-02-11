package units

import "fmt"

type Byte uint64

const (
	TeraBytes Byte = Factor10e3 * GigaBytes
	GigaBytes Byte = Factor10e3 * MegaBytes
	MegaBytes Byte = Factor10e3 * KiloBytes
	KiloBytes Byte = Factor10e3 * Bytes
	Bytes     Byte = 1
)

const Factor10e3 = 1000

func (b Byte) String() string {
	switch b {
	case TeraBytes:
		return "TB"
	case GigaBytes:
		return "GB"
	case MegaBytes:
		return "MB"
	case KiloBytes:
		return "KB"
	case Bytes:
		return "B"
	default:
		return "UNKNOWN"
	}
}

func FormatSize(size int64) string {
	unit := Bytes
	for u := TeraBytes; u >= Bytes; u /= Factor10e3 {
		if size >= int64(u) {
			unit = u
			break
		}
	}
	return fmt.Sprintf("%0.2f%s", float64(size)/float64(unit), unit.String())
}
