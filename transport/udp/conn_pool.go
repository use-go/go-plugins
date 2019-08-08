package udp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"
	"time"

	"github.com/micro/go-micro/util/log"

	"github.com/use-go/go-plugins/transport/utils"
)

//UDP Struct
type UDP struct {
	mp             utils.ConcurrentMap
	OutSessionPool utils.OutSessionPool
	uc             *udpChannel
}

//udpChannel for a udp instance
type udpChannel struct {
	ip          string
	port        int
	UDPListener *net.UDPConn
	Remote      string
	Timeout     int
}

//NewUDPSession create a UDP Socket
func NewUDPSession() *UDP {
	return &UDP{
		OutSessionPool: utils.OutSessionPool{},
		mp:             utils.NewConcurrentMap(),
	}
}

//SendToUDP forward package to another stub
func (s *UDP) SendToUDP(packet []byte, localAddr, srcAddr *net.UDPAddr) (err error) {
	//log.Logf("udp packet revecived:%s,%v", srcAddr, packet)
	dstAddr, err := net.ResolveUDPAddr("udp", s.uc.Remote)
	if err != nil {
		log.Logf("resolve udp addr %s fail  fail,ERR:%s", dstAddr.String(), err)
		return
	}
	clientSrcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	conn, err := net.DialUDP("udp", clientSrcAddr, dstAddr)
	if err != nil {
		log.Logf("connect to udp %s fail,ERR:%s", dstAddr.String(), err)
		return
	}
	conn.SetDeadline(time.Now().Add(time.Millisecond * time.Duration(s.uc.Timeout)))
	_, err = conn.Write(packet)
	if err != nil {
		log.Logf("send udp packet to %s fail,ERR:%s", dstAddr.String(), err)
		return
	}
	buf := make([]byte, 512)
	len, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		log.Logf("read udp response from %s fail ,ERR:%s", dstAddr.String(), err)
		return
	}
	//log.Logf("revecived udp packet from %s , %v", dstAddr.String(), respBody)
	_, err = s.uc.UDPListener.WriteToUDP(buf[0:len], srcAddr)
	if err != nil {
		log.Logf("send udp response to cluster fail ,ERR:%s", err)
		return
	}
	//log.Logf("send udp response to cluster success ,from:%s", dstAddr.String())
	return
}

//GetConn retrive a connection for pool
func (s *UDP) GetConn(connKey string) (conn net.Conn, isNew bool, err error) {
	isNew = !s.mp.Has(connKey)
	var _conn interface{}
	if isNew {
		_conn, err = s.OutSessionPool.Pool.Get()
		if err != nil {
			return nil, false, err
		}
		s.mp.Set(connKey, _conn)
	} else {
		_conn, _ = s.mp.Get(connKey)
	}
	conn = _conn.(net.Conn)
	return
}

//Packet create a packge
func Packet(srcAddr string, packet []byte) []byte {
	addrBytes := []byte(srcAddr)
	addrLength := uint16(len(addrBytes))
	bodyLength := uint16(len(packet))
	pkg := new(bytes.Buffer)
	binary.Write(pkg, binary.LittleEndian, addrLength)
	binary.Write(pkg, binary.LittleEndian, addrBytes)
	binary.Write(pkg, binary.LittleEndian, bodyLength)
	binary.Write(pkg, binary.LittleEndian, packet)
	return pkg.Bytes()
}

//ReadUDPPacket read package from conn
func ReadUDPPacket(conn *net.Conn) (srcAddr string, packet []byte, err error) {
	reader := bufio.NewReader(*conn)
	var addrLength uint16
	var bodyLength uint16
	err = binary.Read(reader, binary.LittleEndian, &addrLength)
	if err != nil {
		return
	}
	_srcAddr := make([]byte, addrLength)
	n, err := reader.Read(_srcAddr)
	if err != nil {
		return
	}
	if n != int(addrLength) {
		return
	}
	srcAddr = string(_srcAddr)

	err = binary.Read(reader, binary.LittleEndian, &bodyLength)
	if err != nil {
		return
	}
	packet = make([]byte, bodyLength)
	n, err = reader.Read(packet)
	if err != nil {
		return
	}
	if n != int(bodyLength) {
		return
	}
	return
}
