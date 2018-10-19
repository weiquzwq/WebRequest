package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	BeginTime time.Time // 开始时间
	SecNum    int       // 秒数

	RQNum int    // 最大并发数，由命令行传入
	Url   string // url，由命令行传入

	userNum    int      // 用户数
)

var users []User

type User struct {
	UserId 		int			 // 用户id
	SBCNum     int           // 并发连接数
	QPSNum     int           // 总请求次数
	RTNum      time.Duration // 响应时间
	RTTNum     time.Duration // 平均响应时间
	SuccessNum int           // 成功次数
	FailNum    int           // 失败次数
	mu         sync.Mutex
}

func (u *User) request(url string) {
	var tb time.Time
	var el time.Duration
	for i := 0;i < u.QPSNum;i++ {
		u.SBCNum++
		go func(u *User) {
			for {
				tb = time.Now()
				_, err := http.Get(Url)
				if err == nil {
					el = time.Since(tb)
					u.mu.Lock() // 上锁
					u.SuccessNum++
					u.RTNum += el
					u.mu.Unlock() // 解锁
				} else {
					u.mu.Lock() // 上锁
					u.FailNum++
					u.mu.Unlock() // 解锁
				}
				time.Sleep(1 * time.Second)
			}
		}(u)
	}
}

func (u *User) show() {
	fmt.Printf("用户id：%d,并发数：%d,请求次数：%d,平均响应时间：%s,成功次数：%d,失败次数：%d\n",
		u.UserId,
		u.SBCNum,
		u.SuccessNum + u.FailNum,
		u.RTNum/(time.Duration(SecNum)*time.Second),
		u.SuccessNum,
		u.FailNum)
}

func showAll(us []User) {
	uLen := len(us)

	var SBCNum     int           // 并发连接数
	var RTNum      time.Duration // 响应时间
	var SuccessNum int           // 成功次数
	var FailNum    int           // 失败次数

	for i := 0;i < uLen;i++ {
		SBCNum += us[i].SBCNum
		SuccessNum += us[i].SuccessNum
		FailNum += us[i].FailNum
		RTNum += us[i].RTNum
		us[i].show()
	}
	fmt.Printf("并发数：%d,请求次数：%d,平均响应时间：%s,成功次数：%d,失败次数：%d\n",
		SBCNum,
		SuccessNum+FailNum,
		RTNum / ((time.Duration(SecNum * uLen) * time.Second)),
		SuccessNum,
		FailNum)
	fmt.Println()
}

func init() {
	if len(os.Args) != 4 {
		log.Fatal("用户数 请求次数 url")
	}
	userNum, _ = strconv.Atoi(os.Args[1])
	RQNum, _ = strconv.Atoi(os.Args[2])
	Url = os.Args[3]
	users = make([]User, userNum)
}

func main() {
	go func() {
		for range time.Tick(2 * time.Second) {
			SecNum += 2
			showAll(users)
		}
	}()
	for range time.Tick(1 * time.Second) {
		requite()
	}
}

func requite() {
	c := make(chan int)
	temp := 0
	for i := 0;i < userNum;i++ {
		if RQNum % userNum != 0 && i < RQNum % userNum {
			temp = 1
		} else {
			temp = 0
		}
		users[i].UserId = i
		users[i].QPSNum = RQNum / userNum + temp
		go users[i].request(Url)
		time.Sleep(45 * time.Millisecond)
	}
	<- c	// 阻塞
}
