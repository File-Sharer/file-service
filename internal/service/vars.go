package service

type level struct {
	maxFileSize int64
	maxSpaceSize int64
}

var (
	levelSpaceSizes = map[uint8]level{
		1: {
			maxFileSize: 1_073_741_824, // 1 GB in bytes
			maxSpaceSize: 8_589_934_592, // 8 GB in bytes
		},
		2: {
			maxFileSize: 8_589_934_592, // 8 GB in bytes
			maxSpaceSize: 17_179_869_184, // 16 GB in bytes
		},
		3: {
			maxFileSize: 8_589_934_592, // 8 GB in bytes
			maxSpaceSize: 34_359_738_368, // 32 GB in bytes
		},
	}
)
