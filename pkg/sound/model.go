package sound

const (
	// FormatFloat32LE is the format of 32 bits floats little endian.
	FormatFloat32LE = 0

	// FormatUnsignedInt8 is the format of 8 bits integers.
	FormatUnsignedInt8 = 1

	//FormatSignedInt16LE is the format of 16 bits integers little endian.
	FormatSignedInt16LE = 2
)

type Note struct {
	// Hz
	Freq int64
	// ms
	Duration int64
}
