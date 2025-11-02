package autorpc

type RPCErrorProvider interface {
	Code() int
	Message() string
	Data() interface{}
}

// errorToRPCError converts a Go error to an RPCError.
// If the error implements RPCErrorProvider, it uses the custom code/message/data.
// Otherwise, it defaults to CodeInternalError with the error message.
func errorToRPCError(err error) *RPCError {
	if rpcErr, ok := err.(RPCErrorProvider); ok {
		data := rpcErr.Data()
		rpcError := &RPCError{
			Code:    rpcErr.Code(),
			Message: rpcErr.Message(),
		}
		if data != nil {
			rpcError.Data = data
		}
		return rpcError
	}

	return &RPCError{
		Code:    CodeInternalError,
		Message: err.Error(),
	}
}
