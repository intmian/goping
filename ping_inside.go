package ping_simple

import (
	"net"
	"time"
)

type Ping_info struct {
	Average  float64
	LostRate float64
}

func Ping_inside_simple(host string, c chan Ping_info) {
	Ping_inside(host, c, 5, 32, 1000, false)
}
func Ping_inside(host string, c chan Ping_info, count int, size int, timeout int64, never_stop bool) {

	starttime := time.Now()
	conn, err := net.DialTimeout("ip4:icmp", host, time.Duration(timeout*1000*1000))

	var seq int16 = 1
	id0, id1 := genidentifier(host)
	const ECHO_REQUEST_HEAD_LEN = 8

	sendN := 0
	recvN := 0
	lostN := 0
	shortT := -1
	longT := -1
	sumT := 0

	for count > 0 || never_stop {
		sendN++
		var msg []byte = make([]byte, size+ECHO_REQUEST_HEAD_LEN)
		msg[0] = 8                        // echo
		msg[1] = 0                        // code 0
		msg[2] = 0                        // checksum
		msg[3] = 0                        // checksum
		msg[4], msg[5] = id0, id1         //identifier[0] identifier[1]
		msg[6], msg[7] = gensequence(seq) //sequence[0], sequence[1]

		length := size + ECHO_REQUEST_HEAD_LEN

		check := checkSum(msg[0:length])
		msg[2] = byte(check >> 8)
		msg[3] = byte(check & 255)

		conn, err = net.DialTimeout("ip:icmp", host, time.Duration(timeout*1000*1000))

		checkError(err)

		starttime = time.Now()
		_ = conn.SetDeadline(starttime.Add(time.Duration(timeout * 1000 * 1000)))
		_, err = conn.Write(msg[0:length])

		const ECHO_REPLY_HEAD_LEN = 20

		var receive []byte = make([]byte, ECHO_REPLY_HEAD_LEN+length)
		n, err := conn.Read(receive)
		_ = n

		var endduration int = int(int64(time.Since(starttime)) / (1000 * 1000))

		sumT += endduration

		if err != nil || receive[ECHO_REPLY_HEAD_LEN+4] != msg[4] || receive[ECHO_REPLY_HEAD_LEN+5] != msg[5] || receive[ECHO_REPLY_HEAD_LEN+6] != msg[6] || receive[ECHO_REPLY_HEAD_LEN+7] != msg[7] || endduration >= int(timeout) || receive[ECHO_REPLY_HEAD_LEN] == 11 {
			lostN++
		} else {
			if shortT == -1 {
				shortT = endduration
			} else if shortT > endduration {
				shortT = endduration
			}
			if longT == -1 {
				longT = endduration
			} else if longT < endduration {
				longT = endduration
			}
			recvN++
		}

		seq++
		count--
	}
	c <- Ping_info{float64(sumT) / float64(sendN), float64(lostN) / float64(sendN)}
}
