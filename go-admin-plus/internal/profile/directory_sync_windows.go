package profile

// Windows does not provide the POSIX directory fsync used after file Sync and Close.
func syncDirectory(string) error {
	return nil
}
