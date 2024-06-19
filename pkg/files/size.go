package files

import "fmt"

type Size int64

func (s Size) String() string {
	unit := Byte

	for u := TeraByte; u > Byte; u /= SizeUnitFactor {
		if int64(s) >= int64(u) {
			unit = u
			break
		}
	}

	return fmt.Sprintf("%0.2f%s", float64(s)/float64(unit), unit.String())
}

func FormatSize(size int64) string {
	return Size(size).String()
}

const SizeUnitFactor = 1e3

type SizeUnit int64

const (
	TeraByte SizeUnit = SizeUnitFactor * GigaByte
	GigaByte SizeUnit = SizeUnitFactor * MegaByte
	MegaByte SizeUnit = SizeUnitFactor * KiloByte
	KiloByte SizeUnit = SizeUnitFactor * Byte
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
