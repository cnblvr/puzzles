// Code generated by "stringer -type=statusCode -linecomment -trimprefix Status -output errors_string.go"; DO NOT EDIT.

package app

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[StatusBadRequest-400]
	_ = x[StatusUnauthorized-401]
	_ = x[StatusMethodNotAllowed-405]
	_ = x[StatusInternalServerError-500]
	_ = x[StatusUnknown-9999]
}

const (
	_statusCode_name_0 = "Bad requestUnauthorized"
	_statusCode_name_1 = "Method not allowed"
	_statusCode_name_2 = "Internal server error"
	_statusCode_name_3 = "Unknown error"
)

var (
	_statusCode_index_0 = [...]uint8{0, 11, 23}
)

func (i statusCode) String() string {
	switch {
	case 400 <= i && i <= 401:
		i -= 400
		return _statusCode_name_0[_statusCode_index_0[i]:_statusCode_index_0[i+1]]
	case i == 405:
		return _statusCode_name_1
	case i == 500:
		return _statusCode_name_2
	case i == 9999:
		return _statusCode_name_3
	default:
		return "statusCode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
