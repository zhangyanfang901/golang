package main
 
import (
    "runtime"
    "fmt"
    "syscall"
    "time"
    "io/ioutil"
    "strconv"
    "strings"
    "bytes"
    "log"
    "os/exec"
    "os"
    "sort"
)

var debugLog *log.Logger
var logFile *os.File

type MemStatus struct {
    All  float64 `json:"all"`
    Used float64 `json:"used"`
    Free float64 `json:"free"`
    Self float64 `json:"self"`
}

type Process struct {
    pid  int
    name string
    cpu  float64
    mem  float64
}
type newProcesslist []*Process  //供Process结构按照cpu排序使用
//调用标准库的sort.Sort必须要先实现Len(),Less(),Swap() 三个方法
func (P newProcesslist) Len() int {
	return len(P)
}
func (P newProcesslist) Less(i, j int) bool {
	return P[i].cpu > P[j].cpu
}
func (P newProcesslist) Swap(i, j int) {
	P[i], P[j] = P[j], P[i]
}


func MemStat() MemStatus {
    //自身占用
    memStat := new(runtime.MemStats)
    runtime.ReadMemStats(memStat)
    mem := MemStatus{}
    mem.Self = float64(memStat.Alloc)/float64(1024*1024)

    //系统占用,仅linux/mac下有效
    //system memory usage
    sysInfo := new(syscall.Sysinfo_t)
    err := syscall.Sysinfo(sysInfo)
    if err == nil {
        mem.All = float64(sysInfo.Totalram)/float64(1024*1024) //* uint64(syscall.Getpagesize())
        mem.Free = float64(sysInfo.Freeram)/float64(1024*1024) //* uint64(syscall.Getpagesize())
        mem.Used = mem.All - mem.Free
    }
    return mem
}


func getCPUSample() (idle, total uint64) {
    contents, err := ioutil.ReadFile("/proc/stat")
    if err != nil {
        return
    }
    lines := strings.Split(string(contents), "\n")
    for _, line := range(lines) {
        fields := strings.Fields(line)
        if fields[0] == "cpu" {
            numFields := len(fields)
            for i := 1; i < numFields; i++ {
                val, err := strconv.ParseUint(fields[i], 10, 64)
                if err != nil {
                    fmt.Println("Error: ", i, fields[i], err)
                }
                total += val // tally up all the numbers to get total ticks
                if i == 4 {  // idle is the 5th field in the cpu line
                    idle = val
                }
            }
            return
        }
    }
    return
}

func getProcessInfo() {
    cmd := exec.Command("ps", "aux")
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        debugLog.Fatal(err)
    }
    processes := make([]*Process, 0)
    for {
        line, err := out.ReadString('\n')
        if err!=nil {
            break;
        }
        tokens := strings.Split(line, " ")
        ft := make([]string, 0)
        for _, t := range(tokens) {
            if t!="" && t!="\t" {
                ft = append(ft, t)
            }
        }
        //debugLog.Println(len(ft), ft)
        //fmt.Printf("%c[1;40;36m%s%c[0m\n", 0x1B, ft, 0x1B)
        pid, err := strconv.Atoi(ft[1])
        if err!=nil {
            continue
        }
        cpu, err := strconv.ParseFloat(ft[2], 64)
        if err!=nil {
            debugLog.Fatal(err)
        }

        mem, err := strconv.ParseFloat(ft[3], 64)
        if err!=nil {
            debugLog.Fatal(err)
        }

        name := strings.Replace(ft[10], "\n", "", -1)  
        processes = append(processes, &Process{pid, name, cpu, mem})
    }
    //按进程的cpu使用率进行排序
    sort.Sort(newProcesslist(processes))  //调用标准库的sort.Sort必须要先实现Len(),Less(),Swap() 三个方法.
    for index, p := range(processes) {
        if index < 5{
            debugLog.Printf("Process %6d  **%25s**  takes %f%% of the CPU and  %f%% of the MEM\n", p.pid, p.name, p.cpu, p.mem)
        }
    }
}


func init()  { 

    fileName := "Info.log"
    logFile,err  := os.Create(fileName)
    if err != nil {
        log.Fatalln("open file error")
    }
    debugLog = log.New(logFile,"[Info]",log.LstdFlags)
    debugLog.Println("A Info message here")
}

func main() {

    //defer logFile.Close()
    for{

        //显示内存信息
        debugLog.Println("\n********************************************************************************")  
        debugLog.Println(time.Now().Format("2006-01-02 15:04:05"))  
        memInfo := MemStat()
        debugLog.Printf("\n %c[1;40;32m%s%c[0m\n\n", 0x1B, "---------------------内存相关信息(单位M)----------------------：", 0x1B)
        debugLog.Printf("%+v", memInfo)
        debugLog.Printf("\n %c[1;40;32m%s%c[0m\n\n", 0x1B, "-------------------------------------------------------------：", 0x1B)

        //显示CPU信息
        idle0, total0 := getCPUSample()
        time.Sleep(3 * time.Second)
        idle1, total1 := getCPUSample()

        idleTicks := float64(idle1 - idle0)
        totalTicks := float64(total1 - total0)
        cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks

        debugLog.Printf("\n %c[1;40;33m%s%c[0m\n\n", 0x1B, "-------------------------CPU相关信息--------------------------：", 0x1B)
        debugLog.Printf("CPU usage is %f%% [busy: %f, total: %f]", cpuUsage, totalTicks-idleTicks, totalTicks)
        debugLog.Printf("\n %c[1;40;33m%s%c[0m\n\n", 0x1B, "-------------------------------------------------------------：", 0x1B)

        //显示各进程信息
        debugLog.Printf("\n %c[1;40;34m%s%c[0m\n\n", 0x1B, "-------------------------占用CPU较高进程-----------------------：", 0x1B)
        getProcessInfo()
        debugLog.Printf("\n %c[1;40;34m%s%c[0m\n\n", 0x1B, "--------------------------------------------------------------：", 0x1B)

        time.Sleep(10 * time.Second)
    }

    
    
}


