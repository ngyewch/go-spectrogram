package spectrogram

func IsPowerOfTwo(n uint) bool {
	return (n != 0) && ((n & (n - 1)) == 0)
}
