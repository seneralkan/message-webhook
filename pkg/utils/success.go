package utils

type HTTPSuccessResponse struct {
	Status    string      `json:"status"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

type HTTPSuccessPaginationResponse struct {
	Status     string      `json:"status"`
	Timestamp  int64       `json:"timestamp"`
	Data       interface{} `json:"data"`
	Pagination interface{} `json:"pagination"`
}

func NewSuccessResponse(data interface{}) HTTPSuccessResponse {
	return HTTPSuccessResponse{
		Status:    "success",
		Timestamp: GetCurrentTimestamp(),
		Data:      data,
	}
}

func NewSuccessPaginationResponse(data, pagination interface{}) HTTPSuccessPaginationResponse {
	return HTTPSuccessPaginationResponse{
		Status:     "success",
		Timestamp:  GetCurrentTimestamp(),
		Data:       data,
		Pagination: pagination,
	}
}
