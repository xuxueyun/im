/**
 * The clientapi file of api current module of xxd.
 *
 * @copyright   Copyright 2009-2017 青岛易软天创网络科技有限公司(QingDao Nature Easy Soft Network Technology Co,LTD, www.cnezsoft.com)
 * @license     ZPL (http://zpl.pub/page/zplv12.html)
 * @author      Archer Peng <pengjiangxiu@cnezsoft.com>
 * @package     api
 * @link        http://www.zentao.net
 */
package api

import (
    "yin/xxd/hyperttp"
    "yin/xxd/util"
    "github.com/patrickmn/go-cache"
    "yin/xxd/models"
    "yin/xxd/orm"
    "strconv"
    "github.com/json-iterator/go"
)

var newline = []byte{'\n'}

// 从客户端发来的登录请求，通过该函数转发到后台服务器进行登录验证
func ChatLogin(clientData ParseData) (*models.User) {
    if x, found := util.Che.Get("user"+strconv.FormatInt(clientData.UserID(),10)); found {
        user := x.(*models.User)
        return user
    }else {
        model,err:=orm.GetUser(clientData.UserID())
        if err==nil{
            util.Che.Set("user"+model.UserID,&model,cache.DefaultExpiration)
        }
        return &model
    }
}
func CheckLogin(userid string)(bool)  {
    if x, found := util.Che.Get("user"+userid); found {
        user := x.(*models.User)
        if user!=nil{
            return true
        }
    }
    return false
}
//获取用户
func GetUser(userid int64) (*models.User) {
    if x, found := util.Che.Get("user"+strconv.FormatInt(userid,10)); found {
        user := x.(*models.User)
        return user
    }else {
        model,err:=orm.GetUser(userid)
        if err==nil{
            util.Che.Set("user"+strconv.FormatInt(userid,10),&model,cache.DefaultExpiration)
        }
        return &model
    }
}
//客户端退出
func ChatLogout(serverName string, userID int64) ([]byte, []int64, error) {
    ranzhiServer, ok := RanzhiServer(serverName)
    if !ok {
        util.LogError().Println("no ranzhi server name")
        return nil, nil, util.Errorf("%s\n", "no ranzhi server name")
    }

    request := []byte(`{"module":"chat","method":"logout","userID":` + util.Int642String(userID) + `}`)
    message, err := aesEncrypt(request, ranzhiServer.RanzhiToken)
    if err != nil {
        util.LogError().Println("aes encrypt error:", err)
        return nil, nil, err
    }

    // 到http服务器请求user get list数据
    r2xMessage, err := hyperttp.RequestInfo(ranzhiServer.RanzhiAddr, message)
    if err != nil {
        util.LogError().Println("hyperttp request info error:", err)
        return nil, nil, err
    }

    // 解析http服务器的数据,返回 ParseData 类型的数据
    parseData, err := ApiParse(r2xMessage, ranzhiServer.RanzhiToken)
    if err != nil {
        util.LogError().Println("api parse error", err)
        return nil, nil, err
    }

    sendUsers := parseData.SendUsers()

    x2cMessage := ApiUnparse(parseData, util.Token)
    if x2cMessage == nil {
        return nil, nil, err
    }

    return x2cMessage, sendUsers, nil

}

//重新登录
func RepeatLogin() []byte {
    repeatLogin := []byte(`{"module":"chat","method":"kickoff","message":"当前账号已在其他地方登录，如果不是本人操作，请及时修改密码"}`)
    //repeatLogin := []byte(`{"module":"chat","method:"kickoff","message":"This account logined in another place."}`)

    //message, err := aesEncrypt(repeatLogin, util.Token)
    //if err != nil {
    //    util.LogError().Println("aes encrypt error:", err)
    //    return nil
    //}

    return repeatLogin
}

//测试登录
func TestLogin() []byte {
    loginData := []byte(`{"result":"success","data":{"id":12,"account":"demo8","realname":"\u6210\u7a0b\u7a0b","avatar":"","role":"hr","dept":0,"status":"online","admin":"no","gender":"f","email":"ccc@demo.com","mobile":"","site":"","phone":""},"sid":"18025976a786ec78194e491e7b790731","module":"chat","method":"login"}`)

    //loginData = append(loginData, newline...)
    message, err := aesEncrypt(loginData, util.Token)
    if err != nil {
        util.LogError().Println("aes encrypt error:", err)
        return nil
    }

    return message
}

// 除登录和退出的数据中转.
func TransitData(clientData []byte, serverName string) ([]byte, []int64, error) {
    ranzhiServer, ok := RanzhiServer(serverName)
    if !ok {
        util.LogError().Println("no ranzhi server name")
        return nil, nil, util.Errorf("%s\n", "no ranzhi server name")
    }

    //交换token
    message, err := SwapToken(clientData, util.Token, ranzhiServer.RanzhiToken)
    if err != nil {
        util.LogError().Println("transit data swap token error:", err)
        return nil, nil, err
    }

    // ranzhi to xxd message
    r2xMessage, err := hyperttp.RequestInfo(ranzhiServer.RanzhiAddr, message)
    if err != nil {
        util.LogError().Println("hyperttp request info error:", err)
        return nil, nil, err
    }

    parseData, err := ApiParse(r2xMessage, ranzhiServer.RanzhiToken)
    if err != nil {
        util.LogError().Println("api parse error:", err)
        return nil, nil, err
    }

    sendUsers := parseData.SendUsers()

    // xxd to client message
    x2cMessage := ApiUnparse(parseData, util.Token)
    if x2cMessage == nil {
        return nil, nil, err
    }

    return x2cMessage, sendUsers, nil
}

//获取用户列表
func UserGetlist(serverName string, userID int64) ([]byte, error) {
    ranzhiServer, ok := RanzhiServer(serverName)
    if !ok {
        util.LogError().Println("no ranzhi server name")
        return nil, util.Errorf("%s\n", "no ranzhi server name")
    }

    // 固定的json格式
  request := []byte(`{"module":"chat","method":"userGetlist","params":[""],"userID":` + util.Int642String(userID) + `}`)

    message, err := aesEncrypt(request, ranzhiServer.RanzhiToken)
    if err != nil {
        util.LogError().Println("aes encrypt error:", err)
        return nil, err
    }

    // 到http服务器请求user get list数据
    retMessage, err := hyperttp.RequestInfo(ranzhiServer.RanzhiAddr, message)
    if err != nil {
        util.LogError().Println("hyperttp request info error:", err)
        return nil, err
    }

    //由于http服务器和客户端的token不一致，所以需要进行交换
    retData, err := SwapToken(retMessage, ranzhiServer.RanzhiToken, util.Token)
    if err != nil {
        util.LogError().Println("user get list swap token error:", err)
        return nil, err
    }

    return retData, nil
}

//用户文件SessionID 作用于文件下载 为适配web版客户端
func UserFileSessionID(serverName string, userID int64) ([]byte , error) {
    sessionID := util.GetMD5( serverName + util.Int642String(userID) + util.Int642String(util.GetUnixTime()) )
    sessionData := []byte(`{"module":"chat","method":"SessionID","sessionID":"` + sessionID + `"}`)

    //将sessionID 存入公共空间
    util.CreateUid(serverName, userID,sessionID)

    sessionData, err := aesEncrypt(sessionData, util.Token)
    if err != nil {
        util.LogError().Println("aes encrypt error:", err)
        return nil, err
    }

    return sessionData , nil
}

//获取用户的会话列表
func Getlist(serverName string, userID int64) ([]byte, error) {
    ranzhiServer, ok := RanzhiServer(serverName)
    if !ok {
        util.LogError().Println("no ranzhi server name")
        return nil, util.Errorf("%s\n", "no ranzhi server name")
    }

    // 固定的json格式
    request := []byte(`{"module":"chat","method":"getlist","userID":` + util.Int642String(userID) + `}`)
    message, err := aesEncrypt(request, ranzhiServer.RanzhiToken)
    if err != nil {
        util.LogError().Println("aes encrypt error:", err)
        return nil, err
    }

    // 到http服务器请求get list数据
    retMessage, err := hyperttp.RequestInfo(ranzhiServer.RanzhiAddr, message)
    if err != nil {
        util.LogError().Println("hyperttp request info error:", err)
        return nil, err
    }

    //由于http服务器和客户端的token不一致，所以需要进行交换
    retData, err := SwapToken(retMessage, ranzhiServer.RanzhiToken, util.Token)
    if err != nil {
        util.LogError().Println("get list swap token error:", err)
        return nil, err
    }

    return retData, nil

}

//获取离线消息
func GetofflineMessages(userID int64) ([]byte,bool) {
    arr:=orm.OfflineMessages(userID)
    var isMsg bool
    if len(arr)>0 {
        isMsg=true
    }else{
        isMsg=false
    }
    //bytes:=[][]byte{}
    //for _,v:= range arr {
    //    msgByte:=[]byte(`{"module":"chat","method":"offmsg","sender":"`+v.Sender+`","recver":"`+v.Sender+`","time":"`+v.Time+`","sendername":"`+v.Sendername+`","recvername":"`+v.Recvername+`","content":"`+v.Msg+`"}`)
    //    bytes=append(bytes, msgByte)
    //}
    var json = jsoniter.ConfigCompatibleWithStandardLibrary
    jsonData,err:=json.Marshal(&arr)
    util.Check(err)

    //ranzhiServer, ok := RanzhiServer(serverName)
    //if !ok {
    //    util.LogError().Println("no ranzhi server name")
    //    return nil, util.Errorf("%s\n", "no ranzhi server name")
    //}
	//
    //// 固定的json格式
    //request := []byte(`{"module":"chat","method":"getOfflineMessages","userID":` + util.Int642String(userID) + `}`)
    //message, err := aesEncrypt(request, ranzhiServer.RanzhiToken)
    //if err != nil {
    //    util.LogError().Println("aes encrypt error:", err)
    //    return nil, err
    //}
	//
    //// 到http服务器请求get list数据
    //retMessage, err := hyperttp.RequestInfo(ranzhiServer.RanzhiAddr, message)
    //if err != nil {
    //    util.LogError().Println("hyperttp request info error:", err)
    //    return nil, err
    //}
	//
    ////由于http服务器和客户端的token不一致，所以需要进行交换
    //retData, err := SwapToken(retMessage, ranzhiServer.RanzhiToken, util.Token)
    //if err != nil {
    //    util.LogError().Println("get off line message swap token error:", err)
    //    return nil, err
    //}

    return jsonData,isMsg
}

// 与客户端间的错误通知
func RetErrorMsg(errCode, errMsg string) ([]byte, error) {
    errApi := `{"module":"chat","method":"error","code":` + errCode + `,"message":"` + errMsg + `"}`
    message, err := aesEncrypt([]byte(errApi), util.Token)
    if err != nil {
        util.LogError().Println("aes encrypt error:", err)
        return nil, err
    }

    return message, nil
}

// 获取然之服务器名称
func RanzhiServer(serverName string) (util.RanzhiServer, bool) {
    if serverName == "" {
        info, ok := util.Config.RanzhiServer[util.Config.DefaultServer]
        return info, ok
    }

    info, ok := util.Config.RanzhiServer[serverName]
    return info, ok
}

//服务器名称
func (pd ParseData) ServerName() string {
    params, ok := pd["params"]
    if !ok {
        return ""
    }

    // api中server name在数组固定位置为0
    ret := params.([]interface{})
    return ret[0].(string)
}

//账号
func (pd ParseData) Account() string {
    params, ok := pd["params"]
    if !ok {
        return ""
    }

    // api中account在数组固定位置为1
    ret := params.([]interface{})
    return ret[1].(string)
}

//密码
func (pd ParseData) Password() string {
    params, ok := pd["params"]
    if !ok {
        return ""
    }

    // api中password在数组固定位置为2
    ret := params.([]interface{})
    return ret[2].(string)
}

//状态
func (pd ParseData) Status() string {
    params, ok := pd["params"]
    if !ok {
        return ""
    }

    // api中status在数组固定位置为3
    ret := params.([]interface{})
    return ret[3].(string)
  }