package trygo

const (
	ERROR_CODE_OK                            = 0
	ERROR_CODE_RUNTIME                       = 1000 //运行时异常
	ERROR_CODE_PARAM_IS_EMPTY                = 1001 //参数为空
	ERROR_CODE_OBJECT_NOT_EXIST              = 1002 //对象不存在
	ERROR_CODE_OBJECT_ALREADY_EXIST          = 1003 //对象已经存在
	ERROR_CODE_PARAM_ILLEGAL                 = 1004 //非法参数
	ERROR_CODE_OPERATE_ILLEGAL               = 1005 //非法操作
	ERROR_CODE_DATABASE_CONNECT_FAILED       = 1100 //连接数据库失败
	ERROR_CODE_DATABASE_CONNECT_CLOSE_FAILED = 1101 //关闭数据库连接失败
	ERROR_CODE_DATABASE_QUERY_FAILED         = 1102 //数据库操作失败
)

var ERROR_INFO_MAP map[int]string

func init() {
	ERROR_INFO_MAP = make(map[int]string)
	ERROR_INFO_MAP[ERROR_CODE_OK] = "ok"
	ERROR_INFO_MAP[ERROR_CODE_RUNTIME] = "runtime exception"
	ERROR_INFO_MAP[ERROR_CODE_PARAM_IS_EMPTY] = "parameter is empty"
	ERROR_INFO_MAP[ERROR_CODE_OBJECT_NOT_EXIST] = "object not exist"
	ERROR_INFO_MAP[ERROR_CODE_OBJECT_ALREADY_EXIST] = "object is exists"
	ERROR_INFO_MAP[ERROR_CODE_PARAM_ILLEGAL] = "illegal parameter"
	ERROR_INFO_MAP[ERROR_CODE_OPERATE_ILLEGAL] = "illegal operation"
	ERROR_INFO_MAP[ERROR_CODE_DATABASE_CONNECT_FAILED] = "connect to database failed"
	ERROR_INFO_MAP[ERROR_CODE_DATABASE_CONNECT_CLOSE_FAILED] = "closing database connection failed"
	ERROR_INFO_MAP[ERROR_CODE_DATABASE_QUERY_FAILED] = "database query failed"

}
