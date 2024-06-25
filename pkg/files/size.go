package files

import "strconv"

const (
	metric = 1000
	binary = 1024
)

const (
	Byte = 1

	KiloByte  = metric * Byte
	MegaByte  = metric * KiloByte
	GigaByte  = metric * MegaByte
	TeraByte  = metric * GigaByte
	PetaByte  = metric * TeraByte
	ExaByte   = metric * PetaByte
	ZettaByte = metric * ExaByte
	YottaByte = metric * ZettaByte

	KibiByte = binary * Byte
	MebiByte = binary * KibiByte
	GibiByte = binary * MebiByte
	TebiByte = binary * GibiByte
	PebiByte = binary * TebiByte
	ExbiByte = binary * PebiByte
	ZebiByte = binary * ExbiByte
	YobiByte = binary * ZebiByte
)

const ShortestLength = -1

func FormatSizeMetricUnits(size float64, precision int) string {
	return formatSize(size, precision, sizeUnit(size, Byte, YottaByte, metric))
}

func FormatSizeBinaryUnits(size float64, precision int) string {
	return formatSize(size, precision, sizeUnit(size, Byte, YobiByte, binary))
}

func formatSize(size float64, precision int, unit float64) string {
	return strconv.FormatFloat(size/unit, 'f', precision, 64) + sizeUnitString(unit)
}

func sizeUnit(size, min, max, factor float64) float64 {
	for u := max; u > min; u /= factor {
		if size >= u {
			return u
		}
	}
	return min
}

func sizeUnitString(size float64) string {
	switch size {
	case Byte:
		return "B"
	case KiloByte:
		return "KB"
	case MegaByte:
		return "MB"
	case GigaByte:
		return "GB"
	case TeraByte:
		return "TB"
	case PetaByte:
		return "PB"
	case ExaByte:
		return "EB"
	case ZettaByte:
		return "ZB"
	case YottaByte:
		return "YB"
	case KibiByte:
		return "KiB"
	case MebiByte:
		return "MiB"
	case GibiByte:
		return "GiB"
	case TebiByte:
		return "TiB"
	case PebiByte:
		return "PiB"
	case ExbiByte:
		return "EiB"
	case ZebiByte:
		return "ZiB"
	case YobiByte:
		return "YiB"
	default:
		return ""
	}
}
