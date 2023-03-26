package errorx

import (
	"fmt"
	"golang.org/x/text/language"
	"runtime"
	"strings"
)

const (
	// 0-999 reserved
	CodeSuccess = 0
	CodeFailed  = 1
	CodeSystem  = 2

	CodeInvalidToken = 401 // token无效
	CodeNotFound     = 404 // 未找到相关资源

	// params
	CodeParamInvalid  = 10001 // 参数错误
	CodeParamRequired = 10002 // 缺少必传参数

	// auth
	CodeAuthBadSignature      = 10101 // 签名错误
	CodeAuthInvalidUserOrPswd = 10102 // 用户名或密码错误
	CodeAuthPermissionDenied  = 10103 // 无权限
	CodeAuthNotFoundUser      = 10104 // 没有找到用户
	CodeIdentityNotFound      = 10105 // 没有找到用户身份

	CodeRealNameVerifyRepeat = 10201 //  已实名，不能重复申请认证
	CodeVerifyCode           = 10202 // 验证码验证失败
	CodeVerifyCodeRate       = 10203 // 验证码请求太频繁
	CodeIdentityRepeat       = 10204 // 用户身份标识重复
	CodeNicknameRepeat       = 10205 // 用户昵称重复
	CodeReferralCode         = 10206 // 邀请码错误
)

var MessagesCN = map[int]string{
	// 0-999 reserved
	CodeSuccess: "处理成功",
	CodeFailed:  "处理失败",
	CodeSystem:  "系统错误",

	CodeInvalidToken: "token无效", // token无效
	CodeNotFound:     "未找到对应资源", // 未找到相关资源

	// params
	CodeParamInvalid:  "参数无效或格式错误", // 参数错误
	CodeParamRequired: "缺少必要参数",    // 缺少必传参数

	// auth
	CodeAuthBadSignature:      "签名错误",      // 签名错误
	CodeAuthInvalidUserOrPswd: "用户名或密码错误",  // 用户名或密码错误
	CodeAuthPermissionDenied:  "无权限",       // 无权限
	CodeAuthNotFoundUser:      "用户不存在",     // 没有找到用户
	CodeIdentityNotFound:      "用户身份信息不存在", // 没有找到用户身份

	CodeRealNameVerifyRepeat: "已实名认证，不可再次提交申请",     //  已实名，不能重复申请认证
	CodeVerifyCode:           "验证码不存在或者无效",         // 验证码验证失败
	CodeVerifyCodeRate:       "验证码请求太频繁，明日再试",      // 验证码请求太频繁
	CodeIdentityRepeat:       "用户身份标识（手机号或者邮箱）已注册", // 用户身份标识重复
	CodeNicknameRepeat:       "用户昵称已注册",            // 用户昵称重复
	CodeReferralCode:         "邀请码无效",              // 邀请码错误
}

var MessagesJP = map[int]string{
	// 0-999 reserved
	CodeSuccess: "success",
	CodeFailed:  "failed",
	CodeSystem:  "system error",

	CodeInvalidToken: "token invalid", // token无效
	CodeNotFound:     "not found",     // 未找到相关资源

	// params
	CodeParamInvalid:  "parameter invalid", // 参数错误
	CodeParamRequired: "parameter require", // 缺少必传参数

	// auth
	CodeAuthBadSignature:      "bad signature",            // 签名错误
	CodeAuthInvalidUserOrPswd: "invalid user or password", // 用户名或密码错误
	CodeAuthPermissionDenied:  "permission denied",        // 无权限
	CodeAuthNotFoundUser:      "not found user",           // 没有找到用户
	CodeIdentityNotFound:      "not found identity",       // 没有找到用户身份

	CodeRealNameVerifyRepeat: "realname verify repeat",              //  已实名，不能重复申请认证
	CodeVerifyCode:           "invalid verification code",           // 验证码验证失败
	CodeVerifyCodeRate:       "verification code request too often", // 验证码请求太频繁
	CodeIdentityRepeat:       "identity registered",                 // 用户身份标识重复
	CodeNicknameRepeat:       "nickname registered",                 // 用户昵称重复
	CodeReferralCode:         "invalid referral code",               // 邀请码错误
}

var Messages = map[int]string{
	// 0-999 reserved
	CodeSuccess: "success",
	CodeFailed:  "failed",
	CodeSystem:  "system error",

	CodeInvalidToken: "token invalid", // token无效
	CodeNotFound:     "not found",     // 未找到相关资源

	// params
	CodeParamInvalid:  "param invalid", // 参数错误
	CodeParamRequired: "param require", // 缺少必传参数

	// auth
	CodeAuthBadSignature:      "bad signature",            // 签名错误
	CodeAuthInvalidUserOrPswd: "invalid user or password", // 用户名或密码错误
	CodeAuthPermissionDenied:  "permission denied",        // 无权限
	CodeAuthNotFoundUser:      "not found user",           // 没有找到用户
	CodeIdentityNotFound:      "not found identity",       // 没有找到用户身份

	CodeRealNameVerifyRepeat: "realname verify repeat",              //  已实名，不能重复申请认证
	CodeVerifyCode:           "invalid verification code",           // 验证码验证失败
	CodeVerifyCodeRate:       "verification code request too often", // 验证码请求太频繁
	CodeIdentityRepeat:       "identity registered",                 // 用户身份标识重复
	CodeNicknameRepeat:       "nickname registered",                 // 用户昵称重复
	CodeReferralCode:         "invalid referral code",               // 邀请码错误
}

var (
	ErrSystem       = NewError(CodeSystem, GetMessage(CodeSystem))
	ErrFailed       = NewError(CodeFailed, GetMessage(CodeFailed))
	ErrNotFound     = NewError(CodeNotFound, GetMessage(CodeNotFound))
	ErrInvalidToken = NewError(CodeInvalidToken, GetMessage(CodeInvalidToken))

	// params
	ErrParamInvalid  = NewError(CodeParamInvalid, GetMessage(CodeParamInvalid))
	ErrParamRequired = NewError(CodeParamRequired, GetMessage(CodeParamRequired))

	// auth
	ErrAuthBadSignature      = NewError(CodeAuthBadSignature, GetMessage(CodeAuthBadSignature))
	ErrAuthInvalidUserOrPswd = NewError(CodeAuthInvalidUserOrPswd, GetMessage(CodeAuthInvalidUserOrPswd))
	ErrAuthPermissionDenied  = NewError(CodeAuthPermissionDenied, GetMessage(CodeAuthPermissionDenied))
	ErrAuthNotFoundUser      = NewError(CodeAuthNotFoundUser, GetMessage(CodeAuthNotFoundUser))
	ErrIdentityNotFound      = NewError(CodeIdentityNotFound, GetMessage(CodeIdentityNotFound))

	ErrRealNameVerifyRepeat = NewError(CodeRealNameVerifyRepeat, GetMessage(CodeRealNameVerifyRepeat))
	ErrVerifyCode           = NewError(CodeVerifyCode, GetMessage(CodeVerifyCode))
	ErrVerifyCodeRate       = NewError(CodeVerifyCodeRate, GetMessage(CodeVerifyCodeRate))
	ErrIdentityRepeat       = NewError(CodeIdentityRepeat, GetMessage(CodeIdentityRepeat))
	ErrNicknameRepeat       = NewError(CodeNicknameRepeat, GetMessage(CodeNicknameRepeat))
	ErrReferralCode         = NewError(CodeReferralCode, GetMessage(CodeReferralCode))
)

type Error struct {
	Code    int      `json:"code" xml:"code"`
	Message string   `json:"message" xml:"message"`
	Stack   []string `json:"-" xml:"-"`
}

func NewError(code int, msg string) Error {
	return Error{Code: code, Message: msg, Stack: make([]string, 0, 1)}
}

func (e Error) Error() string {
	return fmt.Sprintf("code:%d message:%s statck:%v", e.Code, e.Message, e.Stack)
}

func (e Error) MultiErr(err error) Error {
	if err != nil {
		if e1, ok := err.(Error); ok {
			e.Message += ", " + e1.Message
			return e.addStack(e1.Stack...)
		} else {
			e.Message += ", " + err.Error()
		}
	}
	return e
}

func (e Error) MultiMsg(v ...interface{}) Error {
	msg := fmt.Sprint(v...)
	if msg != "" {
		e.Message += ", " + msg
	}
	return e
}

func (e Error) CodeMsg() (int, string) {
	return e.Code, e.Message
}

func (e Error) addStack(v ...string) Error {
	e.Stack = append(e.Stack, v...)
	return e
}

func WithStack(err error) error {
	if err == nil {
		return nil
	}
	frames := runtime.CallersFrames(callers())
	stack := ""
	frame, more := frames.Next()
	if more {
		stack = fmt.Sprintf("\n%s:%d %s", frame.File, frame.Line, frame.Function)
	}

	if e, ok := err.(Error); ok {
		return e.addStack(stack)
	}
	return NewError(CodeFailed, err.Error()).addStack(stack)
}

func callers() []uintptr {
	const depth = 8
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[0:n]
}

// ToSystemError if not Error, to system error
func ToSystemError(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(Error); ok {
		return e
	}
	// to internal
	return NewError(CodeSystem, err.Error())
}

func GetMessage(code int, lang ...string) string {
	if len(lang) == 0 {
		return Messages[code]
	}
	switch {
	case strings.HasPrefix(lang[0], language.Chinese.String()):
		return MessagesCN[code]
	case strings.HasPrefix(lang[0], language.Japanese.String()):
		return MessagesJP[code]
	default:
		return Messages[code]
	}
}
