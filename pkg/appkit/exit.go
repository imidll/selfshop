package appkit

func ExitOnError(
	run func() error, exit func(code int),
) {
	if err := run(); err != nil {
		exit(1)
	}
}
