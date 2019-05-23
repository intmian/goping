// goping 包包含了各种ping及一个动态界面
package goping

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)
// SimplePing 是一个默认参数版的Ping
func SimplePing(host string, c chan int) {
	Ping(host, c, 4, 32, 1000, false)
}
// Ping 可以根据参数在命令行上输出结果
func Ping(host string, c chan int, count int, size int, timeout int64, never_stop bool) {

	cname, _ := net.LookupCNAME(host)
	startTime := time.Now()
	conn, err := net.DialTimeout("ip4:icmp", host, time.Duration(timeout*1000*1000))
	if err != nil {
		panic(err)
	}
	ip := conn.RemoteAddr()
	fmt.Println("正在 Ping " + cname + " [" + ip.String() + "] 具有 32 字节的数据:")

	var seq int16 = 1
	id0, id1 := genIdentifier(host)
	const EchoRequestHeadLen = 8

	sendN := 0
	recvN := 0
	lostN := 0
	shortT := -1
	longT := -1
	sumT := 0

	for count > 0 || never_stop {
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

		sumT += endDuration

		time.Sleep(1000 * 1000 * 1000)

		if err != nil || receive[EchoReplyHeadLen+4] != msg[4] || receive[EchoReplyHeadLen+5] != msg[5] || receive[EchoReplyHeadLen+6] != msg[6] || receive[EchoReplyHeadLen+7] != msg[7] || endDuration >= int(timeout) || receive[EchoReplyHeadLen] == 11 {
			lostN++
			fmt.Println("对 " + cname + "[" + ip.String() + "]" + " 的请求超时。")
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
			ttl := int(receive[8])
			//			fmt.Println(ttl)
			fmt.Println("来自 " + cname + "[" + ip.String() + "]" + " 的回复: 字节=32 时间=" + strconv.Itoa(endDuration) + "ms TTL=" + strconv.Itoa(ttl))
		}

		seq++
		count--
	}
	stat(ip.String(), sendN, lostN, recvN, shortT, longT, sumT)
	c <- 1
}

func checkSum(msg []byte) uint16 {
	sum := 0

	length := len(msg)
	for i := 0; i < length-1; i += 2 {
		sum += int(msg[i])*256 + int(msg[i+1])
	}
	if length%2 == 1 {
		sum += int(msg[length-1]) * 256 // notice here, why *256?
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += sum >> 16
	var answer uint16 = uint16(^sum)
	return answer
}

func checkError(err error) {
	if err != nil {
		_,err := fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	}
}

func genSequence(v int16) (byte, byte) {
	ret1 := byte(v >> 8)
	ret2 := byte(v & 255)
	return ret1, ret2
}

func genIdentifier(host string) (byte, byte) {
	return host[0], host[1]
}

func stat(ip string, sendN int, lostN int, recvN int, shortT int, longT int, sumT int) {
	fmt.Println()
	fmt.Println(ip, " 的 Ping 统计信息:")
	fmt.Printf("    数据包: 已发送 = %d，已接收 = %d，丢失 = %d (%d%% 丢失)，\n", sendN, recvN, lostN, int(lostN*100/sendN))
	fmt.Println("往返行程的估计时间(以毫秒为单位):")
	if recvN != 0 {
		fmt.Printf("    最短 = %dms，最长 = %dms，平均 = %dms\n", shortT, longT, sumT/sendN)
	}
}
