package main

import (
	"actor"
	"base"
	"fmt"
	"message"
	"network"

	"github.com/golang/protobuf/proto"
)

type (
	EventProcess struct {
		actor.Actor

		Client      *network.ClientSocket
		AccountId   int64
		PlayerId    int64
		AccountName string
		SimId       int64
	}

	IEventProcess interface {
		actor.IActor
		LoginGame()
		LoginAccount()
		SendPacket(proto.Message)
	}
)

func SendPacket(packet proto.Message) {
	buff := message.Encode(packet)
	fmt.Println("SendPacket.....buff.....", buff)
	buff = base.SetTcpEnd(buff)
	fmt.Println("len(buff)................", len(buff))
	fmt.Println("SendPacket..SetTcpEnd...buff.....", buff)
	CLIENT.Send(buff)
}

func (this *EventProcess) SendPacket(packet proto.Message) {
	buff := message.Encode(packet)
	fmt.Println("SendPacket.....buff.....", buff)
	buff = base.SetTcpEnd(buff)
	fmt.Println("len(buff)................", len(buff))
	fmt.Println("SendPacket..SetTcpEnd...buff.....", buff)
	this.Client.Send(buff)
}

func (this *EventProcess) PacketFunc(socketid int, buff []byte) bool {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("EventProcess PacketFunc", err)
		}
	}()

	packetId, data := message.Decode(buff)
	packet := message.GetPakcet(packetId)
	if packet == nil {
		return true
	}
	err := proto.Unmarshal(data, packet)
	if err == nil {
		bitstream := base.NewBitStream(make([]byte, 1024), 1024)
		if !message.GetProtoBufPacket(packet, bitstream) {
			return true
		}
		var io actor.CallIO
		io.Buff = bitstream.GetBuffer()
		io.SocketId = socketid
		this.Send(io)
		return true
	}

	return true
}

func (this *EventProcess) Init(num int) {
	this.Actor.Init(num)
	this.RegisterCall("W_C_SelectPlayerResponse", func(packet *message.W_C_SelectPlayerResponse) {
		fmt.Println("client EventProcess ..W_C_SelectPlayerResponse....")
		this.AccountId = *packet.AccountId
		nLen := len(packet.PlayerData)
		fmt.Println(".......W_C_SelectPlayerResponse.....", len(packet.PlayerData), this.AccountId, packet.PlayerData)
		if nLen == 0 {
			packet1 := &message.C_W_CreatePlayerRequest{PacketHead: message.BuildPacketHead(this.AccountId, int(message.SERVICE_WORLDSERVER)),
				PlayerName: proto.String("我是大坏蛋"),
				Sex:        proto.Int32(int32(0))}
			this.SendPacket(packet1)
		} else {
			this.PlayerId = *packet.PlayerData[0].PlayerID
			this.LoginGame()
		}
	})

	this.RegisterCall("W_C_CreatePlayerResponse", func(packet *message.W_C_CreatePlayerResponse) {
		fmt.Println("client EventProcess ..W_C_CreatePlayerResponse....")
		if *packet.Error == 0 {
			this.PlayerId = *packet.PlayerId
			this.LoginGame()
		} else { //创建失败

		}
	})

	this.RegisterCall("A_C_LoginRequest", func(packet *message.A_C_LoginRequest) {
		fmt.Println("client EventProcess ..message.C_A_RegisterRequest....")
		if *packet.Error == base.ACCOUNT_NOEXIST {
			packet1 := &message.C_A_RegisterRequest{PacketHead: message.BuildPacketHead(0, int(message.SERVICE_ACCOUNTSERVER)),
				AccountName: packet.AccountName, Password: packet.Password, SocketId: proto.Int32(0)}
			this.SendPacket(packet1)
		} else if *packet.Error == base.PASSWORD_ERROR {
			fmt.Println("client recieve 账号或密码错误........")
		}
	})

	this.RegisterCall("A_C_RegisterResponse", func(packet *message.A_C_RegisterResponse) {
		fmt.Println("client EventProcess ..A_C_RegisterResponse....")
		//注册失败
		if *packet.Error != 0 {
		}
	})

	this.RegisterCall("W_C_ChatMessage", func(packet *message.W_C_ChatMessage) {
		fmt.Println("client EventProcess ..W_C_ChatMessage....")
		fmt.Println("收到【", *packet.SenderName, "】发送的消息[", *packet.Message+"]")
	})

	this.Actor.Start()
}

func (this *EventProcess) LoginGame() {
	fmt.Println("LoginGame......this.AccountId.....", this.AccountId, "......this.PlayerId.....", this.PlayerId)
	packet1 := &message.C_W_Game_LoginRequset{PacketHead: message.BuildPacketHead(this.AccountId, int(message.SERVICE_WORLDSERVER)),
		PlayerId: proto.Int64(this.PlayerId)}
	this.SendPacket(packet1)
}

var (
	id int
)

func (this *EventProcess) LoginAccount() {
	id++
	//this.AccountName = fmt.Sprintf("test%d", id)
	this.AccountName = fmt.Sprintf("test%d", base.RAND().RandI(0, 7000))
	// this.AccountName = "test2298"
	packet1 := &message.C_A_LoginRequest{PacketHead: message.BuildPacketHead(0, int(message.SERVICE_ACCOUNTSERVER)),
		AccountName: proto.String(this.AccountName), Password: proto.String("123456"), BuildNo: proto.String(base.BUILD_NO), SocketId: proto.Int32(0)}
	fmt.Println("send packet1.....message.SERVICE_ACCOUNTSERVER....", base.BUILD_NO)
	this.SendPacket(packet1)
}

var (
	PACKET *EventProcess
)
