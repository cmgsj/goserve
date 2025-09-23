package files

import "strconv"

const (
	MetricFactor = 1000
	BinaryFactor = 1024
)

const (
	Byte = 1

	KiloByte = MetricFactor * Byte
	MegaByte = MetricFactor * KiloByte
	GigaByte = MetricFactor * MegaByte
	TeraByte = MetricFactor * GigaByte
	PetaByte = MetricFactor * TeraByte
	ExaByte  = MetricFactor * PetaByte

	KibiByte = BinaryFactor * Byte
	MebiByte = BinaryFactor * KibiByte
	GibiByte = BinaryFactor * MebiByte
	TebiByte = BinaryFactor * GibiByte
	PebiByte = BinaryFactor * TebiByte
	ExbiByte = BinaryFactor * PebiByte
)

const ShortestLengthPrecision = -1

func FormatSizeMetric(size float64, precision int) string {
	return formatSize(size, precision, sizeUnit(size, Byte, ExaByte, MetricFactor))
}

func FormatSizeBinary(size float64, precision int) string {
	return formatSize(size, precision, sizeUnit(size, Byte, ExbiByte, BinaryFactor))
}

func formatSize(size float64, precision int, unit float64) string {
	return strconv.FormatFloat(size/unit, 'f', precision, 64) + sizeUnitString(unit)
}

func sizeUnit(size, low, high, factor float64) float64 {
	for u := high; u > low; u /= factor {
		if size >= u {
			return u
		}
	}

	return low
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
	default:
		return ""
	}
}
