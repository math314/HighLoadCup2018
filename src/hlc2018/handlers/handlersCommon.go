package handlers

import "strconv"

type HlcHttpError struct {
	HttpStatusCode int
	Err            error
}

func (e *HlcHttpError) Error() string {
	return "status: " + strconv.Itoa(e.HttpStatusCode) + ", error: " + e.Err.Error()
}
