// Code generated by "stringer -linecomment -type compatTestCaseResultType"; DO NOT EDIT.

package integration

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[nonEmptyResult-0]
	_ = x[emptyResult-1]
}

const _compatTestCaseResultType_name = "nonEmptyResultemptyResult"

var _compatTestCaseResultType_index = [...]uint8{0, 14, 25}

func (i compatTestCaseResultType) String() string {
	if i < 0 || i >= compatTestCaseResultType(len(_compatTestCaseResultType_index)-1) {
		return "compatTestCaseResultType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _compatTestCaseResultType_name[_compatTestCaseResultType_index[i]:_compatTestCaseResultType_index[i+1]]
}
