package cairo

type ErrorStatus Status

func (e ErrorStatus) Error() string {
	return StatusToString(Status(e))
}
