package netgate

import (
	"actor"
	"base"
	"fmt"
	"strings"
)

type (
	UserServerProcess struct {
		actor.Actor
	}
)

func (this *UserServerProcess) Init(num int) {
	this.Actor.Init(num)
	this.RegisterCall("DISCONNECT", func(socketId int) {
		SERVER.GetPlayerMgr().SendMsg("DEL_ACCOUNT", socketId)
	})

	this.Actor.Start()
}

func (this *UserServerProcess) PacketFunc(id int, buff []byte) bool {
	/*packetId,_ := message.Decode(buff)
	packet := message.GetPakcet(packetId)
	if packet != nil{
		return false
	}*/

	// fmt.Println("UserServerProcess....PacketFunc.....")
	var io actor.CallIO
	io.Buff = buff
	io.SocketId = id

	bitstream := base.NewBitStream(io.Buff, len(io.Buff))
	funcName := bitstream.ReadString()
	fmt.Println("UserServerProcess...funcName....................", funcName)
	funcName = strings.ToLower(funcName)
	pFunc := this.FindCall(funcName)
	fmt.Println("UserServerProcess...pFunc....................", pFunc)
	if pFunc != nil {
		this.Send(io)
		return true
	}

	return false
}
