package trygo

type SessionInterface interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	Delete(key string)
	SessionID() string
	Flush() error //delete all data
}

type ProviderInterface interface {
	SessionInit(gclifetime int64, config string) error
	SessionRead(sid string) (SessionInterface, error)
	SessionExist(sid string) bool
	SessionRegenerate(oldsid, sid string) (SessionInterface, error)
	SessionDestroy(sid string) error
	SessionAll() int //get all active session
	SessionGC()
}

//type sessionManagerInterface interface {
//	SessionStart() SessionInterface
//	SetHashFunc(hf hashFunc)
//	Destroy(HTTPCookieHandlerInterface)
//	RegenerateSession(cookieHandler HTTPCookieHandlerInterface) SessionInterface
//	GC()
//}
