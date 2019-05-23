package goping

import (
	"net"
	"time"
)
/*
PingInfo is
type PingInfo struct {
	Average  float32
	LostRate float32
}
*/
type PingInfo struct {
	Average  float32
	LostRate float32
}
// PingInsideSimple 是默认参数版的PingInside
func PingInsideSimple(host string, c chan PingInfo) {
	PingInside(host, c, 4, 32, 1000, false)
}
// PingInside 是可以通过PingInfo返回结果的Ping 但是不能返回每次的结果
func PingInside(host string, c chan PingInfo, count int, size int, timeout int64, neverStop bool) {

	startTime := time.Now()
	conn, err := net.DialTimeout("ip4:icmp", host, time.Duration(timeout*1000*1000))

	var seq int16 = 1
	id0, id1 := genIdentifier(host)
	const EchoRequestHeadLen = 8

	sendN := 0
	recvN := 0
	lostN := 0
	shortT := -1
	longT := -1
	sumT := 0

	for count > 0 || neverStop {
		sendN++
		var msg []byte = make([]byte, size+EchoRequestHeadLen)
		msg[0] = 8                        // echo
		msg[1] = 0                        // code 0
		msg[2] = 0                        // checksum
		msg[3] = 0                        // checksum
		msg[4], msg[5] = id0, id1         //identifier[0] identifier[1]
		msg[6], msg[7] = genSequence(seq) //sequence[0], sequence[1]

		length := size + EchoRequestHeadLen

		check := checkSum(msg[0:length])
		msg[2] = byte(check >> 8)
		msg[3] = byte(check & 255)

		conn, err = net.DialTimeout("ip:icmp", host, time.Duration(timeout*1000*1000))

		checkError(err)

		startTime = time.Now()
		_ = conn.SetDeadline(startTime.Add(time.Duration(timeout * 1000 * 1000)))
		_, err = conn.Write(msg[0:length])

		const EchoReplyHeadLen = 20

		var receive []byte = make([]byte, EchoReplyHeadLen+length)
		n, err := conn.Read(receive)
		_ = n

		var endDuration int = int(int64(time.Since(startTime)) / (1000 * 1000))
		if int64(endDuration) < timeout {
			sumT += endDuration
		}
		if err != nil || receive[EchoReplyHeadLen+4] != msg[4] || receive[EchoReplyHeadLen+5] != msg[5] || receive[EchoReplyHeadLen+6] != msg[6] || receive[EchoReplyHeadLen+7] != msg[7] || endDuration >= int(timeout) || receive[EchoReplyHeadLen] == 11 {
			lostN++
		} else {
			if shortT == -1 {
				shortT = endDuration
			} else if shortT > endDuration {
				shortT = endDuration
			}
			if longT == -1 {
				longT = endDuration
			} else if longT < endDuration {
				longT = endDuration
			}
			recvN++
		}

		seq++
		count--
	}
	if lostN == sendN {
		c <- PingInfo{float32(timeout), 1}
	} else {
		c <- PingInfo{float32(sumT) / float32(recvN), float32(lostN) / float32(sendN)}
		// 除去丢失的算时间
	}
}
