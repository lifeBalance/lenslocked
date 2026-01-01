package apperrors

type publicError struct {
	err error
	msg string
}

// Add the Error() method so that it implements error interface.
func (pe publicError) Error() string {
	return pe.err.Error()
}

// Convert an error to a public error
func Public(err error, msg string) error {
	return publicError{err, msg}
}

func (pe publicError) Public() string {
	return pe.msg // return the public message
}

func (pe publicError) Unwrap() error {
	return pe.err
}
