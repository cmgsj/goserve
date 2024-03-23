package files

import "fmt"

const factor = 1e3

type SizeUnit uint64

const (
	TeraByte SizeUnit = factor * GigaByte
	GigaByte SizeUnit = factor * MegaByte
	MegaByte SizeUnit = factor * KiloByte
	KiloByte SizeUnit = factor * Byte
	Byte     SizeUnit = 1
)

func (u SizeUnit) String() string {
	switch u {
	case TeraByte:
		return "TB"
	case GigaByte:
		return "GB"
	case MegaByte:
		return "MB"
	case KiloByte:
		return "KB"
	default:
		return "B"
	}
}

func FormatSize(size int64) string {
	unit := Byte
	for u := TeraByte; u > Byte; u /= factor {
		if size >= int64(u) {
			unit = u
			break
		}
	}
	return fmt.Sprintf("%0.2f%s", float64(size)/float64(unit), unit.String())
}
