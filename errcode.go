package ssss

const (
	ERROR_CODE_RUNTIME                       int = 1000 //运行时异常
	ERROR_CODE_PARAM_IS_EMPTY                int = 1001 //参数为空
	ERROR_CODE_OBJECT_NOT_EXIST              int = 1002 //对象不存在
	ERROR_CODE_OBJECT_ALREADY_EXIST          int = 1003 //对象已经存在
	ERROR_CODE_PARAM_ILLEGAL                 int = 1004 //非法参数
	ERROR_CODE_OPERATE_ILLEGAL               int = 1005 //非法操作
	ERROR_CODE_DATABASE_CONNECT_FAILED       int = 1100 //连接数据库失败
	ERROR_CODE_DATABASE_CONNECT_CLOSE_FAILED int = 1101 //关闭数据库连接失败
	ERROR_CODE_DATABASE_QUERY_FAILED         int = 1102 //数据库操作失败
)

var ERROR_INFO_MAP map[int]string

func init() {
	ERROR_INFO_MAP = make(map[int]string)
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
