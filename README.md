# go simple ping

## 简介

1. 提供可视化的网络实时状况展示
    ![可视化](https://i.loli.net/2019/05/22/5ce4daad433a386721.png)
    ![](https://i.loli.net/2019/05/22/5ce4e2f47de7460950.png)
    ![](https://i.loli.net/2019/05/22/5ce4e30c65a5732204.png)
2. gui界面提供灵活的软件配置
   - 可通过json配置测试主机列表
   - 可通过json配置各项参数
3. 提供一大堆go实现高效ping（支持多个ping并发）

## 安装

- 需要调用ping接口或改造gui

  使用`go get`工具进行安装源码

    ```shell
    go get github.com/intmian/ping_go_simple
    ```
- 只需gui工具的话，下载gui可执行文件[go gui](https://github.com/intmian/ping_go_simple/releases/download/v1.0-alpha/ping.rar)

## 使用

### 引入

```go
import "github.com/intmian/ping_simple"
```

### 接口

- `Ping(host string, c chan int, count int, size int, timeout int64, never_stop bool)`

  - `host` 主机号

  - `count` 一次发送几个包

  - `size` 包大小

  - `timeout` 时间

  - `never_stop` 是否为永久的ping（一直ping，直到ctrl + c终止）

  - ```go
    done chan bool
    go Ping(host, done, count, timeout, never_stop)
    <-done
    ```

  - 会输出标准的ping信息（和系统自带的ping一样）

- `SimplePing(host string, c chan int)`

  - ```go
    func SimplePing(host string, c chan int) {
    	Ping(host, c, 4, 32, 1000, false)
    }
    ```

- `PingInfo`

  - ```
    type PingInfo struct {
    	Average  float32
    	LostRate float32
    }
    ```

- `Ping_inside(host string, c chan PingInfo, count int, size int, timeout int64, never_stop bool)`

  - 和之前的ping一样，不过数据以PingInfo形式输出

  - ```go
    data chan bool
    go Ping_inside(host, data, count, timeout, never_stop)
    temp := <-data
    print(temp.Average)
    print(temp.LostRate)
    ```

- `Ping_inside_simple(host string, c chan PingInfo)`

  - ```go
    func Ping_inside_simple(host string, c chan PingInfo) {
    	Ping_inside(host, c, 4, 32, 1000, false)
    }
    ```

- `Gui()`

  - 以gui形式显示网络状况，从运行目录中读取 `setting.json` 与 `hosts.json` 中的配置。ctrl + c 结束后保存图片到根目录的 `avg.png` 与 `lost_rate.png`  

  - ```go
    Gui()
    ```

  - 配置文件样本

    - `setting.json`

      ```json
      {
        "sleepTime" : 10,
        "repaintTime" : 0.5,
        "count" : 5
      }
      ```

    - `hosts.json`

      ```json
      [
      	"www.baidu.com",
      	"www.intmian.com"
      ]
      ```
