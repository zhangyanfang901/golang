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
)

var debugLog *log.Logger
var logFile *os.File

type MemStatus struct {
    All  uint64 `json:"all"`
    Used uint64 `json:"used"`
    Free uint64 `json:"free"`
    Self uint64 `json:"self"`
}

type Process struct {
    pid int
    cpu float64
    mem float64
}

func MemStat() MemStatus {
    //自身占用
    memStat := new(runtime.MemStats)
    runtime.ReadMemStats(memStat)
    mem := MemStatus{}
    mem.Self = memStat.Alloc

    //系统占用,仅linux/mac下有效
    //system memory usage
    sysInfo := new(syscall.Sysinfo_t)
    err := syscall.Sysinfo(sysInfo)
    if err == nil {
        mem.All = sysInfo.Totalram //* uint64(syscall.Getpagesize())
        mem.Free = sysInfo.Freeram //* uint64(syscall.Getpagesize())
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
        processes = append(processes, &Process{pid, cpu, mem})
    }
    for _, p := range(processes) {
        debugLog.Printf("Process %8d takes %f%% of the CPU and  %f%% of the MEM\n", p.pid, p.cpu, p.mem)
    }
}


func init()  { 

    fileName := "Info.log"
    logFile,err  := os.Create(fileName)
    if err != nil {
        log.Fatalln("open file error")
    }
    debugLog = log.New(logFile,"[Info]",log.Llongfile)
    debugLog.Println("A Info message here")
}

func main() {

    //defer logFile.Close()
    for{

        debugLog.Println("\n********************************************************************************")  
        debugLog.Println(time.Now().Format("2006-01-02 15:04:05"))  
        memInfo := MemStat()
        debugLog.Printf("\n %c[1;40;32m%s%c[0m\n\n", 0x1B, "-------------------------内存相关信息-------------------------：", 0x1B)
        debugLog.Printf("%+v", memInfo)
        debugLog.Printf("\n %c[1;40;32m%s%c[0m\n\n", 0x1B, "-------------------------------------------------------------：", 0x1B)

        idle0, total0 := getCPUSample()
        time.Sleep(3 * time.Second)
        idle1, total1 := getCPUSample()

        idleTicks := float64(idle1 - idle0)
        totalTicks := float64(total1 - total0)
        cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks

        debugLog.Printf("\n %c[1;40;33m%s%c[0m\n\n", 0x1B, "-------------------------CPU相关信息--------------------------：", 0x1B)
        debugLog.Printf("CPU usage is %f%% [busy: %f, total: %f]", cpuUsage, totalTicks-idleTicks, totalTicks)
        debugLog.Printf("\n %c[1;40;33m%s%c[0m\n\n", 0x1B, "-------------------------------------------------------------：", 0x1B)

        debugLog.Printf("\n %c[1;40;34m%s%c[0m\n\n", 0x1B, "-------------------------各进程相关信息-----------------------：", 0x1B)
        getProcessInfo()
        debugLog.Printf("\n %c[1;40;34m%s%c[0m\n\n", 0x1B, "--------------------------------------------------------------：", 0x1B)

        time.Sleep(6 * time.Second)
    }

    
    
}


