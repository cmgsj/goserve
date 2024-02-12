package bytes

import "fmt"

const Factor10e3 = 1000

type Unit uint64

const (
	TeraByte Unit = Factor10e3 * GigaByte
	GigaByte Unit = Factor10e3 * MegaByte
	MegaByte Unit = Factor10e3 * KiloByte
	KiloByte Unit = Factor10e3 * Byte
	Byte     Unit = 1
)

func (u Unit) String() string {
	switch u {
	case TeraByte:
		return "TB"
	case GigaByte:
		return "GB"
	case MegaByte:
		return "MB"
	case KiloByte:
		return "KB"
	case Byte:
		return "B"
	default:
		return "UNKNOWN"
	}
}

func FormatSize(size int64) string {
	unit := Byte
	for u := TeraByte; u > Byte; u /= Factor10e3 {
		if size >= int64(u) {
			unit = u
			break
		}
	}
	return fmt.Sprintf("%0.2f%s", float64(size)/float64(unit), unit.String())
}
