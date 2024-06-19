package files

import "strconv"

const SizeFactor = 1e3

type Size int64

const (
	TeraByte Size = SizeFactor * GigaByte
	GigaByte Size = SizeFactor * MegaByte
	MegaByte Size = SizeFactor * KiloByte
	KiloByte Size = SizeFactor * Byte
	Byte     Size = 1
)

func (s Size) Unit() Size {
	unit := Byte

	for u := TeraByte; u > Byte; u /= SizeFactor {
		if s >= u {
			unit = u
			break
		}
	}

	return unit
}

func (s Size) UnitString() string {
	switch s {
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

func (s Size) String() string {
	return s.Format(-1)
}

func (s Size) Format(precision int) string {
	unit := s.Unit()
	return strconv.FormatFloat(float64(s)/float64(unit), 'f', precision, 64) + s.UnitString()
}

func FormatSize(size int64, precision int) string {
	return Size(size).Format(precision)
}
