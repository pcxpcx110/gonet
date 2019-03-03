package network_test

import (
	"fmt"
	"base"
	"testing"
)

var(
	m_pInBuffer []byte
	m_pInBuffer1 []byte
)

func Test66(t *testing.T)  {
	for i :=0;i < 1000000;i++{
		buff := []byte{}
		for j := 0; j < 1; j++{
			buff = append(buff, SetTcpEnd(base.GetPacket("test1",[200]int64{1,2,3,4,5,6}))...)
		}
		ReceivePacket(0,buff)
	}
}

func Test2(t *testing.T)  {

	for i :=0;i < 1000000;i++{
		buff := []byte{}
		for j := 0; j < 1; j++{
			buff = append(buff, SetTcpEnd1(base.GetPacket("test1", [200]int64{1,2,3,4,5,6}))...)
		}
		ReceivePacket1(0,buff)
	}
}

func SetTcpEnd(buff []byte) []byte{
	buff = append(base.IntToBytes(len(buff)), buff...)
	return buff
}

func SetTcpEnd1(buff []byte) []byte{
	buff = append(buff, []byte(base.TCP_END)...)
	return buff
}

func ReceivePacket(Id int, dat []byte){
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("ReceivePacket", err) // 接受包错误
		}
	}()
	//找包结束
	seekToTcpEnd := func(buff []byte) (bool, int){
		nLen := len(buff)
		if nLen < base.PACKET_HEAD_SIZE{
			return false, 0
		}

		nSize := base.BytesToInt(buff[0:4])
		if nSize + base.PACKET_HEAD_SIZE <= nLen{
			return true, nSize+base.PACKET_HEAD_SIZE
		}
		return false, 0
	}

	buff := append(m_pInBuffer, dat...)
	m_pInBuffer = []byte{}
	nCurSize := 0
	//fmt.Println(this.m_pInBuffer)
ParsePacekt:
	nPacketSize := 0
	nBufferSize := len(buff[nCurSize:])
	bFindFlag := false
	bFindFlag, nPacketSize = seekToTcpEnd(buff[nCurSize:])
	//fmt.Println(bFindFlag, nPacketSize, nBufferSize)
	if bFindFlag{
		if nBufferSize == nPacketSize{		//完整包
			//this.HandlePacket(Id, buff[nCurSize+base.PACKET_HEAD_SIZE:nCurSize+nPacketSize])
			nCurSize += nPacketSize
		}else if ( nBufferSize > nPacketSize){
			//this.HandlePacket(Id, buff[nCurSize+base.PACKET_HEAD_SIZE:nCurSize+nPacketSize])
			nCurSize += nPacketSize
			goto ParsePacekt
		}
	}else if nBufferSize < base.MAX_PACKET{
		m_pInBuffer = buff[nCurSize:]
	}else{
		fmt.Println("超出最大包限制，丢弃该包")
	}
}

func  ReceivePacket1(Id int, dat []byte){
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("ReceivePacket1", err) // 接受包错误
		}
	}()
	//找包结束
	seekToTcpEnd := func(buff []byte) (bool, int){
		nLen := len(buff)
		for	i := 0; i < nLen - 1; i++ {
			if (buff[i] == base.TCP_END[0] && buff[i+1] == base.TCP_END[1]){
				return true, i+2
			}
		}
		return false, 0
	}

	buff := append(m_pInBuffer1, dat...)
	m_pInBuffer1 = []byte{}
	nCurSize := 0
	//fmt.Println(this.m_pInBuffer)
ParsePacekt:
	nPacketSize := 0
	nBufferSize := len(buff[nCurSize:])
	bFindFlag := false
	bFindFlag, nPacketSize = seekToTcpEnd(buff[nCurSize:])
	//fmt.Println(bFindFlag, nPacketSize, nBufferSize)
	if bFindFlag{
		if nBufferSize == nPacketSize{		//完整包
			nCurSize += nPacketSize
		}else if ( nBufferSize > nPacketSize){
			nCurSize += nPacketSize
			goto ParsePacekt
		}
	}else if nBufferSize < base.MAX_PACKET{
		m_pInBuffer1 = buff[nCurSize:]
	}else{
		fmt.Println("超出最大包限制，丢弃该包")
	}
}
