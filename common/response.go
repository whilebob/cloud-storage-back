package common

// ResponseType 统一响应结构
type ResponseType struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success 创建成功响应
func Success(code int, message string, data interface{}) ResponseType {
	return ResponseType{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// Error 创建错误响应
func Error(code int, message string) ResponseType {
	return ResponseType{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// SuccessWithData 创建带数据的成功响应
func SuccessWithData(data interface{}) ResponseType {
	return Success(200, "success", data)
}

// ErrorWithCode 创建带自定义错误码的错误响应
func ErrorWithCode(code int, message string) ResponseType {
	return Error(code, message)
}
