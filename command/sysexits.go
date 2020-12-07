package command

// Based on sysexits.h from FreeBSD (also used by Linux)
// https://www.freebsd.org/cgi/man.cgi?query=sysexits&manpath=FreeBSD+13.0-current

var (
	// The command executed successfully.
	ExitOK = 0

	// The command was used incorrectly, e.g., with the wrong number of
	// arguments, a bad flag, a bad syntax in a parameter, or whatever.
	ExitUsage = 64

	// The input data was incorrect in some way. This should only be used for
	// user's data and not system files.
	ExitDataError = 65

	// An input file (not a system file) did not exist or was not readable. This
	// could also include errors like ``No message'' to a mailer (if it cared to
	// catch it).
	ExitNoInput = 66

	// The user specified did not exist. This might be used for mail addresses
	// or remote logins.
	ExitNoUser = 67

	// The host specified did not exist. This is used in mail addresses or
	// network requests.
	ExitNoHost = 68

	// A service is unavailable. This can occur if a support program or file
	// does not exist. This can also be used as a catchall message when
	// something you wanted to do doesn't work, but you don't know why.
	ExitUnavailable = 69

	// An internal software error has been detected. This should be limited to
	// non-operating system related errors as possible.
	ExitSoftware = 70

	// An operating system error has been detected. This is intended to be used
	// for such things as ``cannot fork'', ``cannot create pipe'', or the like.
	// It includes things like getuid returning a user that does not exist in
	// the passwd file.
	ExitOSError = 71

	// Some system file (e.g., /etc/passwd, /var/run/utmp, etc.) does not exist,
	// cannot be opened, or has some sort of error (e.g., syntax error).
	ExitOSFile = 72

	// A (user specified) output file cannot be created.
	ExitCantCreate = 73

	// An error occurred while doing I/O on some file.
	ExitIOError = 74

	// Temporary failure, indicating something that is not really an error. In
	// sendmail, this means that a mailer (e.g.) could not create a connection,
	// and the request should be reattempted later.
	ExitTemporaryFailure = 75

	// The remote system returned something that was ``not possible'' during a
	// protocol exchange.
	ExitProtocol = 76

	// You did not have sufficient permission to perform the operation. This is
	// not intended for file system problems, which should use ExitNoInput or
	// ExitCantCreate, but rather for higher level permissions.
	ExitNoPermission = 77

	// Something was found in an unconfigured or misconfigured state.
	ExitConfig = 78
)

type CommandErrorer interface {
	error
	ExitStatus() int
}

type UsageError struct{ Message string }

func (e *UsageError) Error() string   { return e.Message }
func (e *UsageError) ExitStatus() int { return ExitUsage }
