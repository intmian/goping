package ping_simple

import (
	"encoding/json"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"image/color"
	"io/ioutil"
	"math/rand"
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
	time     float32 // 时间梭
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
	//time.Sleep(50 * time.Second) // debug
	endSignal <- true
}

func printer(clock <-chan bool, endSig <-chan bool, pingData <-chan pingData) map[string]PingDataSmall {
	/*
		输出数据
		clock : 周期输出时钟信号
		endSig: 外部的终止信号
		pingData: 外部来的数据
	*/
	clear()
	strFlash := []string{"↑↑↑运行中↑↑↑", "→→→运行中→→→",
		"↓↓↓运行中↓↓↓", "←←←运行中←←←"}
	strBack := "\b\b\b\b\b\b\b\b\b\b\b\b\b\b\b\b\b\b" // 需要多一点\b 不知道为什么和方向箭头一样数量的话在我的cmder里面就显示不清楚了

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
				fmt.Printf("%-20s : %6.2fms %3.0f%%\n", k, v.avgTime, v.lostRate*100)
			}
			print(strFlash[i])
			i++
			i = i % 4
		}
	}
}
func pinger(pingSig chan<- pingData, hosts []string, endSig <-chan bool, sleepTime float32, count int) {
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
				go Ping_inside(hostC.host, hostC.c, count, 32, 3000, false)
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
func bindChanToChansBool(chanSource chan bool, chansTarget []chan bool) {
	/*
		克隆chanSource处的信号，到各个chanTarget中
	*/
	for {
		sigTemp := <-chanSource
		for _, sigC := range chansTarget {
			sigC <- sigTemp
		}
	}
}
func bindChanToChansPing(chanSource chan pingData, chansTarget ...chan pingData) {
	/*
		克隆chanSource处的信号，到各个chanTarget中
	*/
	for {
		sigTemp := <-chanSource
		for _, sigC := range chansTarget {
			sigC <- sigTemp

		}
	}
}
func guiExec(endChan chan bool, pingDataChanPro chan pingData) []PingTimeData {
	var pingDataAll []PingTimeData
	for { // exec
		select {
		case <-endChan: // 退出
			return pingDataAll
		case temp := <-pingDataChanPro: // 同步的存一下
			pingDataAll = append(pingDataAll,
				PingTimeData{float32(time.Now().Hour()) + float32(time.Now().Minute())/60 + float32(time.Now().Second())/3600, temp.host, temp.avgTime, temp.lostRate})
		}
	}
}

func Gui() {
	rand.Seed(time.Now().UnixNano())
	data, err := ioutil.ReadFile("setting.json")
	if err != nil {
		fmt.Printf("文件%s不存在\n", "setting.json")
		return
	}
	settingData := map[string]float32{}
	_ = json.Unmarshal(data, &settingData) // 反序列化
	sleepTime := settingData["sleepTime"]
	repaintTime := settingData["repaintTime"]
	count := int(settingData["count"])

	data2, err2 := ioutil.ReadFile("hosts.json")
	if err2 != nil {
		fmt.Printf("文件%s不存在\n", "setting.json")
		return
	}
	var hosts []string
	_ = json.Unmarshal(data2, &hosts) // 反序列化

	endChanOrigin := make(chan bool) // 发出终止信号
	const serviceNum = 4             // 除了各个goroutine的中止信号外，还有一个gui自己用的处于 [0]
	endChans := make([]chan bool, 0) // 接受终止信号
	clockChan := make(chan bool)
	pingDataChanOri := make(chan pingData, 20)    // ping数据来源
	pingDataChanTarget := make(chan pingData, 20) // 送向printer的数据
	pingDataChanPro := make(chan pingData, 20)    // 送向本地储存的数据

	for i := 0; i < serviceNum; i++ {
		endChans = append(endChans, make(chan bool)) // 方便无阻塞的发送信号给各个协程
	}
	// 绑定一下多播的信号
	go bindChanToChansPing(pingDataChanOri, pingDataChanTarget, pingDataChanPro)
	go bindChanToChansBool(endChanOrigin, endChans)

	endChan := endChans[0] // 调度用的
	go clock(clockChan, repaintTime, endChans[1])
	go systemSignal(endChanOrigin)
	go printer(clockChan, endChans[2], pingDataChanTarget)
	go pinger(pingDataChanOri, hosts, endChans[3], sleepTime, count)

	pingDataAll := guiExec(endChan, pingDataChanPro) // 存档的ping数据

	n := len(hosts)
	avgs := make([][]float32, n)
	lostRates := make([][]float32, n)
	times := make([][]float32, n) // 将时间化为h.m/64+s/3600
	for _, data := range pingDataAll {
		for i, name := range hosts {
			if data.host == name {
				avgs[i] = append(avgs[i], data.avgTime)
				lostRates[i] = append(lostRates[i], data.lostRate)
				times[i] = append(times[i], data.time)
			}
		}
	}

	avgPts := make([]plotter.XYs, n) // 平均时间的数据
	for i := 0; i < n; i++ {
		length := len(avgs[i])
		avgPts[i] = make(plotter.XYs, length)
		for j := 0; j < length; j++ {
			avgPts[i][j].X = float64(times[i][j])
			avgPts[i][j].Y = float64(avgs[i][j])
		}
	}
	pAvg, err := plot.New() // 图表
	if err != nil {
		panic(err)
	}
	pAvg.Title.Text = "avg-time"
	pAvg.X.Label.Text = "time(h)"
	pAvg.Y.Label.Text = "time(ms)"
	for i := 0; i < n; i++ {
		l, _ := plotter.NewLine(avgPts[i])
		l.LineStyle.Color = color.RGBA{R: uint8(rand.Intn(256)), G: uint8(rand.Intn(256)), B: uint8(rand.Intn(256)), A: 255}
		pAvg.Add(l) // 颜色是随机填充的
		pAvg.Legend.Add(hosts[i], l)
	}
	_ = pAvg.Save(10*vg.Inch, 10*vg.Inch, "avg.png")
	// 和上面相同
	lostRatesPts := make([]plotter.XYs, n)
	for i := 0; i < n; i++ {
		length := len(lostRates[i])
		lostRatesPts[i] = make(plotter.XYs, length)
		for j := 0; j < length; j++ {
			lostRatesPts[i][j].X = float64(times[i][j])
			lostRatesPts[i][j].Y = float64(lostRates[i][j])
		}
	}
	pLostRate, _ := plot.New()

	pLostRate.Title.Text = "lost rate-time"
	pLostRate.X.Label.Text = "time(h)"
	pLostRate.Y.Label.Text = "lost rate"
	for i := 0; i < n; i++ {
		l, _ := plotter.NewLine(lostRatesPts[i])
		l.LineStyle.Color = color.RGBA{R: uint8(rand.Intn(256)), G: uint8(rand.Intn(256)), B: uint8(rand.Intn(256)), A: 255}
		pLostRate.Add(l)
		pLostRate.Legend.Add(hosts[i], l)
	}
	_ = pLostRate.Save(10*vg.Inch, 10*vg.Inch, "lost_rate.png")
}
