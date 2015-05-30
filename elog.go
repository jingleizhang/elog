package elog

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	LOG_SHIFT_BY_SIZE = 1 //roll log file with fixed size
	LOG_SHIFT_BY_MIN  = 2 //one file every minute
	LOG_SHIFT_BY_HOUR = 3 //one file every hour
	LOG_SHIFT_BY_DAY  = 4 //one file every day

	LOG_WRITE_INTERVAL_MSEC    = 10     //millisecond
	LOG_WRITE_BUFFER_CHECK_LEN = 1 * 32 //KByte

	LOG_FATAL = 0x01 //Log Level: VIP / fatal /error / info / debug
	LOG_ERROR = 0x02
	LOG_INFO  = 0x04
	LOG_DEBUG = 0x08
)

const (
	TRACKING_URL        = "http://tracking.bdp.cn/stat?guid=tracking&"
	LOG_JSON_IP         = "src_ip"
	LOG_JSON_DATE       = "src_date"
	LOG_JSON_LEVEL      = "log_level"
	LOG_JSON_NOTE       = "log_note"
	LOG_JSON_ENDPOINT   = "end_point" //URL Encode
	LOG_JSON_PRIORITY   = "priority"
	LOG_JSON_NEEDFILTER = "needfilter"
	LOG_JSON_METHOD     = "httpmethod"
	LOG_JSON_QUERY      = "query" //URL Encode
	LOG_JSON_VIP        = "vip"
	LOG_JSON_EXTRACT    = "needextractlinks"
)

type ELog struct {
	shiftMode      int
	fileName       string
	prefix         string
	maxSize        int
	logLevel       int
	buffers        []string
	lock           *sync.RWMutex
	dateStr        string //date
	hourStr        string //hour
	minStr         string //minute
	lastWriteMS    int64  //milli second
	keepInTracking bool   //is tracking, default: false
	keepInFile     bool   //is kept in file, default: true
	//keepInConsole  bool   //is kept in console
	trackingChan chan string
	httpClient   *http.Client
}

func NewELog(logprefix string, logmode int, maxSizeKB int, level int) *ELog {
	elog := &ELog{
		shiftMode:      logmode,
		prefix:         logprefix,
		lastWriteMS:    0,
		lock:           &sync.RWMutex{},
		maxSize:        maxSizeKB * 1024,
		logLevel:       level,
		keepInTracking: false,
		keepInFile:     true,
		trackingChan:   make(chan string, 256),
		httpClient:     newHttpClient(10),
	}

	elog.dateStr = GetDateStr()
	elog.hourStr, elog.minStr = GetHourMinuteStr()
	elog.fileName = makeFileName(elog.prefix, elog.dateStr, elog.hourStr, elog.minStr, elog.shiftMode)

	go elog.initTimer()

	return elog
}

func (elog *ELog) DLogFatal(format interface{}, v ...interface{}) {
	if (elog.logLevel & LOG_FATAL) > 0 {
		elog.dlog(elog.getExtraInfo("FATAL") + fmt.Sprint(format) + fmt.Sprint(v...) + "\n")
		elog.flush()
		os.Exit(-1)
	}
}

func (elog *ELog) DLogVIP(format interface{}, v ...interface{}) {
	//VIP: do not check level
	elog.dlog(elog.getExtraInfo("VIP") + fmt.Sprint(format) + fmt.Sprint(v...) + "\n")
}

func (elog *ELog) DLogError(format interface{}, v ...interface{}) {
	if (elog.logLevel & LOG_ERROR) > 0 {
		elog.dlog(elog.getExtraInfo("ERROR") + fmt.Sprint(format) + fmt.Sprint(v...) + "\n")
	}
}

func (elog *ELog) DLogInfo(format interface{}, v ...interface{}) {
	if (elog.logLevel & LOG_INFO) > 0 {
		elog.dlog(elog.getExtraInfo("INFO") + fmt.Sprint(format) + fmt.Sprint(v...) + "\n")
	}
}

func (elog *ELog) DLogDebug(format interface{}, v ...interface{}) {
	if (elog.logLevel & LOG_DEBUG) > 0 {
		elog.dlog(elog.getExtraInfo("DEBUG") + fmt.Sprint(format) + fmt.Sprint(v...) + "\n")
	}
}

func (elog *ELog) Fini() {
	elog.DLogFlush()
	elog.keepInFile = false
	elog.keepInTracking = false
	close(elog.trackingChan)
}

func (elog *ELog) DLogFlush() {
	elog.flush()
}

func (elog *ELog) SetTracking(b bool, moduleName int) {
	if b && !elog.keepInTracking {
		go elog.tracking(moduleName)
	}

	elog.keepInTracking = b
}

func (elog *ELog) SetKeptInFile(b bool) {
	elog.keepInFile = b
}

func newHttpClient(timeOutSeconds int) *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				timeout := time.Duration(timeOutSeconds) * time.Second
				deadline := time.Now().Add(timeout)
				c, err := net.DialTimeout(netw, addr, timeout)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
			DisableKeepAlives: true,
		},
	}
	return client
}

func makeFileName(prefix, date, hour, min string, logMode int) string {
	var fileName string
	switch logMode {
	case LOG_SHIFT_BY_MIN:
		fileName = prefix + "_" + date + "_" + hour + "_" + min
	case LOG_SHIFT_BY_HOUR:
		fileName = prefix + "_" + date + "_" + hour
	case LOG_SHIFT_BY_DAY:
		fileName = prefix + "_" + date
	case LOG_SHIFT_BY_SIZE:
		fileName = prefix
	default:
		fileName = prefix
	}
	fileName += "." + GetLocalIp()
	return fileName
}

func (elog *ELog) getExtraInfo(level string) string {
	ip := GetLocalIp()
	_, file, line, ok := runtime.Caller(2)
	if ok {
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				file = file[i+1:]
				break
			}
		}
	} else {
		file = "unknown"
		line = 0
	}
	return fmt.Sprintf("%s|%s|%s:%d|%s|", ip, time.Now().Format("2006/01/02 15:04:05"), file, line, level)
}

func getURLFormat(nodeType int, logRaw string) string {
	if len(logRaw) > 0 {
		jsonStr := ""
		tks := strings.Split(logRaw, "|")
		if len(tks) >= 5 {
			jsonStr += LOG_JSON_IP + "=" + tks[0] + "&"
			jsonStr += LOG_JSON_DATE + "=" + url.QueryEscape(tks[1]) + "&"
			jsonStr += LOG_JSON_LEVEL + "=" + tks[3] + "&"
			jsonStr += tks[4]
			return jsonStr
		}
	}
	return ""
}

func (elog *ELog) shift() {
	if !elog.keepInFile || elog.bufferLength() == 0 {
		return
	}

	date := GetDateStr()
	hour, min := GetHourMinuteStr()

	switch elog.shiftMode {
	case LOG_SHIFT_BY_MIN:
		if elog.minStr != min {
			elog.dateStr = date
			elog.hourStr = hour
			elog.minStr = min
			elog.fileName = makeFileName(elog.prefix, elog.dateStr, elog.hourStr, elog.minStr, elog.shiftMode)
		}
	case LOG_SHIFT_BY_HOUR:
		if elog.hourStr != hour {
			elog.dateStr = date
			elog.hourStr = hour
			elog.minStr = min
			elog.fileName = makeFileName(elog.prefix, elog.dateStr, elog.hourStr, elog.minStr, elog.shiftMode)
		}
	case LOG_SHIFT_BY_DAY:
		if elog.dateStr != date {
			elog.dateStr = date
			elog.hourStr = hour
			elog.minStr = min
			elog.fileName = makeFileName(elog.prefix, elog.dateStr, elog.hourStr, elog.minStr, elog.shiftMode)
		}
	case LOG_SHIFT_BY_SIZE:
		fi, err := os.Stat(elog.fileName)
		if err != nil {
			os.Create(elog.fileName)
		} else if fi.Size() >= (int64)(elog.maxSize) {
			os.Rename(elog.fileName, "."+elog.fileName)
			os.Create(elog.fileName)
			go func() {
				os.Remove("." + elog.fileName)
			}()
		}
	}
}

func (elog *ELog) timerFlush() {
	elog.dlog("")
}

func (elog *ELog) initTimer() {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			elog.timerFlush()
		}
	}
}

func (elog *ELog) bufferLength() int {
	length := 0
	for _, buf := range elog.buffers {
		length += len(buf)
	}
	return length
}

func GetDateStr() string {
	now := time.Now()
	year, mon, day := now.UTC().Date()
	return fmt.Sprintf("%04d%02d%02d", year, mon, day)
}

func GetHourMinuteStr() (string, string) {
	hour := time.Now().Hour()
	minute := time.Now().Minute()

	return fmt.Sprintf("%02d", hour), fmt.Sprintf("%02d", minute)
}

func GetLocalIp() string {
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, addr := range addrs {
			if strings.Contains(addr.String(), "10.") || strings.Contains(addr.String(), "192") || strings.Contains(addr.String(), "172") {
				start := strings.Index(addr.String(), "/")
				if start > 0 {
					return addr.String()[:start]
				}
				return addr.String()
			}
		}
	}
	return "unknown_ip"
}

func (elog *ELog) tracking(nodeType int) {
	for {
		select {
		case entry := <-elog.trackingChan:
			if len(entry) > 0 {
				urlStr := TRACKING_URL + "&" + getURLFormat(nodeType, entry)
				reqest, _ := http.NewRequest("GET", urlStr, nil)
				response, _ := elog.httpClient.Do(reqest)
				if response != nil {
					response.Body.Close()
				}
			}
		}
	}
}

func (elog *ELog) flush() {
	if elog.bufferLength() > 0 {
		if elog.keepInFile {
			logFile, err := os.OpenFile(elog.fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				fmt.Println("Error|" + err.Error())
			} else {
				var buf string
				for _, v := range elog.buffers {
					buf += v
				}
				logFile.WriteString(buf)
			}
			logFile.Close()
			elog.lastWriteMS = time.Now().UnixNano() / 1000000
		}

		if elog.keepInTracking {
			for _, buf := range elog.buffers {
				elog.trackingChan <- buf
			}
		}
		elog.buffers = elog.buffers[:0]
	}
}

func (elog *ELog) dlog(v string) {
	elog.shift()

	//Add data to local buffer, then batch flush to disk.
	elog.lock.Lock()
	elog.buffers = append(elog.buffers, v)
	if ((time.Now().UnixNano()/1000000)-elog.lastWriteMS > LOG_WRITE_INTERVAL_MSEC) || (elog.bufferLength() > LOG_WRITE_BUFFER_CHECK_LEN) {
		elog.flush()
	}
	elog.lock.Unlock()
}
