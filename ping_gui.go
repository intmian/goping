package ping_simple

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type pingData struct {
	host     string  // 主机地址
	avgTime  float32 // 平均往返时间
	lostRate float32 // 丢失率
}
type PingDataSmall struct {
	avgTime  float32 // 平均往返时间
	lostRate float32 // 丢失率
}

type PingTimeData struct {
	time     int64   // 时间梭
	host     string  // 主机地址
	avgTime  float32 // 平均往返时间
	lostRate float32 // 丢失率
}

func clear() {
	/*
		清屏
	*/
	c := exec.Command("cmd", "/c", "cls") //可以根据自己的需要修改参数，自己试试，我也不清楚
	c.Stdout = os.Stdout
	c.Run()
}
func clock(clockSignals chan<- bool, duration float32, endSignals <-chan bool) {
	/*
		每隔一段时间输出一个信号直到结束
		clockSignals: 周期输出时钟信号
		duration: time(s)
		endSignals: 外部的终止信号
	*/
	ticker := time.NewTicker((time.Duration(int64(1000 * duration))) * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			clockSignals <- true
		case <-endSignals:
			return
		}
	}
}
func systemSignal(endSignal chan<- bool) {
	/*
		捕获系统信号，并输出终止信号
		endSignal: 终止信号通道
	*/
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	endSignal <- true
}

func printer(clock <-chan bool, endSig <-chan bool, pingData <-chan pingData) map[string]PingDataSmall {
	/*
		输出数据
		clock : 周期输出时钟信号
		endSig: 外部的终止信号
		pingData: 外部来的数据
	*/
	strFlash := []string{"↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑", "→→→→→→→→→→→→→→→", "↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓", "←←←←←←←←←←←←←←←"}
	strBack := "\b\b\b\b\b\b\b\b\b\b\b\b\b\b\b"
	hostData := make(map[string]PingDataSmall) // 主机号对应的数据

	i := 0
	for {
		select {
		case <-clock: // 刷新状态栏
			print(strBack)
			print(strFlash[i])
			i++
			i = i % 4
		case <-endSig:
			return hostData
		case data := <-pingData: // 刷新全部
			hostData[data.host] = PingDataSmall{data.avgTime, data.lostRate}
			clear()
			for k, v := range hostData {
				fmt.Printf("%20s : %3.2fms %3.2f%%", k, v.avgTime, v.lostRate)
				print(strFlash[i])
				i++
				i = i % 4
			}
		}
	}
}
func pinger(pingSig chan<- pingData, hosts []string, endSig <-chan bool, sleepTime float32) {
	/*
		pingSig 对外的ping信号
	*/
	type hostChan struct {
		host string
		c    chan PingInfo
	}

	var hostCs []hostChan // 包含了域名和并发控制的通道
	for _, host := range hosts {
		hostCs = append(hostCs, hostChan{host, make(chan PingInfo)})
	}

	for {
		select {
		case <-endSig:
			return
		default:
			for _, hostC := range hostCs {
				go Ping_inside(hostC.host, hostC.c, 20, 32, 3000, false)
			} // 运行所有ping
			for _, hostC := range hostCs {
				var temp pingData
				data := <-hostC.c

				temp.host = hostC.host
				temp.lostRate = data.LostRate
				temp.avgTime = data.Average
				pingSig <- temp
			} // 获得所有ping结果并发送出去
		}

		time.Sleep(time.Duration(1000*sleepTime) * time.Millisecond)
	}
}
func bindChanToChans(chanSource chan bool, chansTarget ...chan bool) {
	/*
		克隆chanSource处的信号，到各个chanTarget中
	*/
	sigTemp := <-chanSource
	for _, sigC := range chansTarget {
		sigC <- sigTemp
	}
}
func gui(hosts []string, sleepTime float32, repaintTime float32) {
	endChanOrigin := make(chan bool) // 发出终止信号
	const serviceNum = 4             // 除了各个goroutine的中止信号外，还有一个gui自己用的处于 [0]
	endChans := make([]chan bool, 0) // 接受终止信号
	clockChan := make(chan bool)
	pingDataChanOri := make(chan pingData)    // ping数据来源
	pingDataChanTarget := make(chan pingData) // 送向printer的数据
	pingDatachanPro := make(chan pingData)    // 送向本地储存的数据

	bindChanToChans(endChanOrigin, endChans...)

	for i := 0; i < serviceNum; i++ {
		endChans = append(endChans, make(chan bool))
	}
	endChan := endChans[0] // 调度用的
	go clock(clockChan, repaintTime, endChans[1])
	go systemSignal(endChanOrigin)
	go printer(clockChan, endChans[2], pingDataChanOri)
	go pinger(pingDataChanTarget, hosts, endChans[3], sleepTime)

	pingDatas := make([]PingTimeData, 0) // 存档的ping数据
	for {
		select {
		case <-endChan:
			break;
		case temp := <-pingDataChanOri:
			pingDatas = append(pingDatas, PingTimeData{time.Now().Second()})
		}
	}

	// TODO 处理数据
}
