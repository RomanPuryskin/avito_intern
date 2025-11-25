package errs

const (
	USER_EXISTS     = "USER_EXISTS"
	TEAM_EXISTS     = "TEAM_EXISTS"
	PR_EXISTS       = "PR_EXISTS"
	PR_MERGED       = "PR_MERGED"
	NOT_ASSIGNED    = "NOT_ASSIGNED"
	NO_CANDIDATE    = "NO_CANDIDATE"
	NOT_FOUND       = "NOT_FOUND"
	INTERNAL_SERVER = "INTERNAL_SERVER"
	INVALID_INPUT   = "INVALID_INPUT"
)

type ResponceError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (

	// INVALID_INPUT
	ErrorInvaidInput = ResponceError{
		Code:    INVALID_INPUT,
		Message: "invalid input",
	}

	ErrorInvaidInputFormat = ResponceError{
		Code:    INVALID_INPUT,
		Message: "invalid input format",
	}

	// USER_EXISTS
	ErrorUserAlreadyExists = ResponceError{
		Code:    USER_EXISTS,
		Message: "user with user_id already exists",
	}

	ErrorUserAlreadyExistsByUserName = ResponceError{
		Code:    USER_EXISTS,
		Message: "user with username already exists",
	}

	// NOT_FOUND
	ErrorUserNotFound = ResponceError{
		Code:    NOT_FOUND,
		Message: "user not found",
	}

	ErrorTeamNotFound = ResponceError{
		Code:    NOT_FOUND,
		Message: "team not found",
	}

	ErrorPRNotFound = ResponceError{
		Code:    NOT_FOUND,
		Message: "pr not found",
	}

	// TEAM_EXISTS
	ErrorTeamAlreadyExists = ResponceError{
		Code:    TEAM_EXISTS,
		Message: "team already exists",
	}

	// INTERNAL_SERVER
	ErrorInternal = ResponceError{
		Code:    INTERNAL_SERVER,
		Message: "internal server error",
	}

	// PR_EXISTS
	ErrorPRAlreadyExists = ResponceError{
		Code:    PR_EXISTS,
		Message: "pr already exists",
	}

	// PR_MERGED
	ErrorPRMerged = ResponceError{
		Code:    PR_MERGED,
		Message: "cannot reassign on merged PR",
	}

	// NO_CANDIDATE
	ErrorNoCandidateToReassign = ResponceError{
		Code:    NO_CANDIDATE,
		Message: "no active replacement candidate in team",
	}

	// NOT_ASSIGNED
	ErrorUserNotAssigned = ResponceError{
		Code:    NOT_ASSIGNED,
		Message: "reviewer is not assigned to this PR",
	}
)
