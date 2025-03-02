package utils

type ErrorManager struct{}

type ManagedErrorFunc func() error

func ReturnFirstError(funcs ...ManagedErrorFunc) error {
	for _, fn := range funcs {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
