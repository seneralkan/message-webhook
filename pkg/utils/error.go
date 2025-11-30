package utils

import "context"

const (
	ValidationErrCode = "validation_failed"
	UnexpectedErrCode = "unexpected_error"
	BodyParserErrCode = "body_parser_failed"

	UnexpectedMsg = "An unexpected error has occurred."
	ValidationMsg = "The given data was invalid."
	BodyParserMsg = "The given values could not be parsed."
)

type Error struct {
	Code    string `json:"code"`
	Reason  error  `json:"reason"`
	Message string `json:"message"`
}

func (e Error) GetCode() string {
	return e.Code
}

func (e Error) Error() string {
	return e.Reason.Error()
}

func (e Error) GetMessage() string {
	return e.Message
}

const (
	ErrorStatus = "error"
)

type HTTPErrorResponse struct {
	Status    string      `json:"status"`
	Timestamp int64       `json:"timestamp"`
	Error     ErrorSchema `json:"error"`
}

type ErrorSchema struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

type HTTPErrorResponseWithReason struct {
	Status    string                `json:"status"`
	Timestamp int64                 `json:"timestamp"`
	Error     ErrorSchemaWithReason `json:"error"`
}

type ErrorSchemaWithReason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

type HTTPValidationErrorResponse struct {
	Error     ErrorSchema   `json:"error"`
	Fields    []ErrorFields `json:"fields"`
	Status    string        `json:"status"`
	Timestamp int64         `json:"timestamp"`
}

type ErrorFields struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func NewErrorResponse(ctx context.Context, err error) HTTPErrorResponse {
	schema := ErrorSchema{
		Code:    UnexpectedErrCode,
		Message: UnexpectedMsg,
	}

	if errorBag, ok := err.(Error); ok {
		schema.Code = errorBag.GetCode()
		schema.Message = errorBag.GetMessage()
	}

	return HTTPErrorResponse{
		Error:     schema,
		Status:    ErrorStatus,
		Timestamp: GetCurrentTimestamp(),
	}
}

func NewValidationErrorResponse(errors map[string]string) HTTPValidationErrorResponse {
	var attrs []ErrorFields
	for k, v := range errors {
		attrs = append(attrs, ErrorFields{
			Field:   k,
			Message: v,
		})
	}

	return HTTPValidationErrorResponse{
		Error: ErrorSchema{
			Code:    ValidationErrCode,
			Message: ValidationMsg,
		},
		Fields:    attrs,
		Status:    ErrorStatus,
		Timestamp: GetCurrentTimestamp(),
	}
}

func NewBodyParserErrorResponse() HTTPErrorResponse {
	return HTTPErrorResponse{
		Error: ErrorSchema{
			Code:    BodyParserErrCode,
			Message: BodyParserMsg,
		},
		Status:    ErrorStatus,
		Timestamp: GetCurrentTimestamp(),
	}
}
