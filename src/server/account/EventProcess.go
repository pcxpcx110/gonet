package account

import (
	"actor"
	"base"
	"database/sql"
	"db"
	"fmt"
	"log"
	"message"
	"server/common"

	"github.com/golang/protobuf/proto"
)

type (
	EventProcess struct {
		actor.Actor
		m_db *sql.DB
	}

	IEventProcess interface {
		actor.IActor
	}
)

func (this *EventProcess) Init(num int) {
	this.Actor.Init(num)
	this.m_db = SERVER.GetDB()
	this.RegisterCall("COMMON_RegisterRequest", func(nType int, Ip string, Port int) {
		pServerInfo := new(common.ServerInfo)
		pServerInfo.SocketId = this.GetSocketId()
		pServerInfo.Type = nType
		pServerInfo.Ip = Ip
		pServerInfo.Port = Port

		SERVER.GetServerMgr().SendMsg("CONNECT", pServerInfo)

		switch pServerInfo.Type {
		case int(message.SERVICE_GATESERVER):
			SERVER.GetServer().SendMsgByID(this.GetSocketId(), "COMMON_RegisterResponse")
		case int(message.SERVICE_WORLDSERVER):
			SERVER.GetServer().SendMsgByID(this.GetSocketId(), "COMMON_RegisterResponse")
		}
	})

	//链接断开
	this.RegisterCall("DISCONNECT", func(socketid int) {
		SERVER.GetServerMgr().SendMsg("DISCONNECT", socketid)
	})

	//创建账号
	this.RegisterCall("C_A_RegisterRequest", func(packet *message.C_A_RegisterRequest) {
		fmt.Println("account EventProcess C_A_RegisterRequest...")
		accountName := *packet.AccountName
		password := *packet.Password
		// password := "123456"
		socketId := int(*packet.SocketId)
		Error := 1
		var result string
		var accountId int64
		rows, err := this.m_db.Query(fmt.Sprintf("call `usp_activeaccount`('%s', '%s', %d)", accountName, password, base.UUID.UUID()))
		if err == nil && rows != nil {
			if rows.NextResultSet() {
				rs := db.Query(rows)
				if rs.Next() {
					accountId = rs.Row().Int64("@accountId")
					result = rs.Row().String("@result")
					if result == "0000" {
						SERVER.GetLog().Printf("帐号[%s]创建成功", accountName)
						//登录账号
						SERVER.GetAccountMgr().SendMsg("Account_Login", accountName, accountId, socketId, this.GetSocketId())
						Error = 0
					}
				}
			}
		}

		if Error != 0 {
			SendToClient(this.GetSocketId(), &message.A_C_RegisterResponse{
				PacketHead: message.BuildPacketHead(accountId, 0),
				Error:      proto.Int32(int32(Error)),
				SocketId:   packet.SocketId,
			})
		}
	})

	//登录账号
	this.RegisterCall("C_A_LoginRequest", func(packet *message.C_A_LoginRequest) {
		accountName := *packet.AccountName
		password := *packet.Password
		// password := "543232"
		buildVersion := *packet.BuildNo
		socketId := int(*packet.SocketId)
		error := base.NONE_ERROR

		if base.CVERSION().IsAcceptableBuildVersion(buildVersion) {
			log.Printf("账号[%s]登陆账号服务器", accountName)
			rows, err := this.m_db.Query(fmt.Sprintf("call `usp_login`('%s', '%s')", accountName, password))
			if err == nil && rows != nil {
				if rows.NextResultSet() { //存储过程反馈多个select的时候
					rs := db.Query(rows)
					if rs.Next() {
						accountId := rs.Row().Int64("@accountId")
						result := rs.Row().String("@result")
						fmt.Println("C_A_LoginRequest......query...acountId..", accountId, "..result....", result)
						//register account
						if result == "0001" {
							error = base.ACCOUNT_NOEXIST
						} else if result == "0000" {
							error = base.NONE_ERROR
							SERVER.GetAccountMgr().SendMsg("Account_Login", accountName, accountId, socketId, this.GetSocketId())
						} else if result == "0002" {
							error = base.PASSWORD_ERROR
							// SERVER.GetAccountMgr().SendMsg("ReLoginAccount", accountName, accountId, socketId, this.GetSocketId())
						}
					}
				}
			}
		} else {
			error = base.VERSION_ERROR
			log.Println("版本验证错误 clientVersion=%s,err=%d", buildVersion, error)
		}

		if error != base.NONE_ERROR {
			fmt.Println("返回登陆错误信息.....", error)
			SendToClient(this.GetSocketId(), &message.A_C_LoginRequest{
				PacketHead:  message.BuildPacketHead(0, 0),
				Error:       proto.Int32(int32(error)),
				SocketId:    packet.SocketId,
				AccountName: packet.AccountName,
				Password:    packet.Password,
			})
		}
	})

	//创建玩家
	this.RegisterCall("W_A_CreatePlayer", func(accountId int64, playername string, sex int32, socketId int) {
		rows, err := this.m_db.Query(fmt.Sprintf("call `usp_createplayer`(%d, '%s', %d)", accountId, playername, base.UUID.UUID()))
		if err == nil && rows != nil {
			rs := db.Query(rows)
			if rs.Next() {
				err := rs.Row().Int("@err")
				playerId := rs.Row().Int64("@playerId")
				if err == 0 && playerId > 0 {
					SERVER.GetServer().SendMsgByID(this.GetSocketId(), "A_W_CreatePlayer", accountId, playerId, playername, sex, socketId)
				}
			}
		}
	})

	//删除玩家
	this.RegisterCall("W_A_DeletePlayer", func(accountId int64, playerId int64) {
		this.m_db.Exec(fmt.Sprintf("update tbl_player set delete_flag = 1 where account_id =%d and player_id=%d", accountId, playerId))
	})

	this.Actor.Start()
}
