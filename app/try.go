package app

// Try is a basic function to quickly handling errors during handling POCs.
func Try(err error) {
	if err != nil {
		panic(err)
	}
}
