package jobs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	log "github.com/go-admin-team/go-admin-core/v2/logger"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

var timeFormat = "2006-01-02 15:04:05"
var retryCount = 3

var jobList map[string]JobExec

var cronContexts sync.Map

//var lock sync.Mutex

type JobCore struct {
	InvokeTarget   string
	Name           string
	JobId          int
	EntryId        int
	CronExpression string
	Args           string
	ctx            context.Context
}

// HttpJob 任务类型 http
type HttpJob struct {
	JobCore
}

type ExecJob struct {
	JobCore
}

func (e *ExecJob) Run() {
	ctx := e.context()
	if ctx.Err() != nil {
		return
	}
	startTime := time.Now()
	var obj = jobList[e.InvokeTarget]
	if obj == nil {
		log.Warn("[Job] ExecJob Run job nil")
		return
	}
	var err error
	if contextual, ok := obj.(ContextJobExec); ok {
		err = contextual.ExecContext(ctx, e.Args)
	} else {
		err = CallExec(obj.(JobExec), e.Args)
	}
	if err != nil {
		// 如果失败暂停一段时间重试
		fmt.Println(time.Now().Format(timeFormat), " [ERROR] mission failed! ", err)
	}
	// 结束时间
	endTime := time.Now()

	// 执行时间
	latencyTime := endTime.Sub(startTime)
	//TODO: 待完善部分
	//str := time.Now().Format(timeFormat) + " [INFO] JobCore " + string(e.EntryId) + "exec success , spend :" + latencyTime.String()
	//ws.SendAll(str)
	log.Infof("[Job] JobCore %s exec success , spend :%v", e.Name, latencyTime)
	return
}

// Run http 任务接口
func (h *HttpJob) Run() {
	ctx := h.context()
	startTime := time.Now()
	for attempt := 0; attempt < retryCount; attempt++ {
		if _, err := getWithContext(ctx, h.InvokeTarget); err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Warnf("[Job] mission failed! %v", err)
			if attempt == retryCount-1 {
				return
			}
			delay := time.Duration(attempt+1) * 5 * time.Second
			log.Warnf("[Job] Retry after the task fails %s", delay)
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
			case <-ctx.Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				return
			}
			continue
		}
		break
	}
	// 结束时间
	endTime := time.Now()

	// 执行时间
	latencyTime := endTime.Sub(startTime)
	//TODO: 待完善部分

	log.Infof("[Job] JobCore %s exec success , spend :%v", h.Name, latencyTime)
	return
}

func (c *JobCore) withContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *JobCore) context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

func getWithContext(ctx context.Context, target string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: 30 * time.Second}).Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	return string(contents), err
}

// Setup is the expand-phase compatibility entry point. Callers own the
// returned Module and must stop it during host shutdown.
func Setup(databases map[string]*gorm.DB) (*Module, error) {
	module := NewModuleWithDatabases(databases)
	if err := module.Start(context.Background()); err != nil {
		return nil, err
	}
	return module, nil
}

// AddJob 添加任务 AddJob(invokeTarget string, jobId int, jobName string, cronExpression string)
func AddJob(c *cron.Cron, job Job) (int, error) {
	if job == nil {
		fmt.Println("unknown")
		return 0, nil
	}
	if ctx, exists := cronContexts.Load(c); exists {
		if contextual, ok := job.(interface{ withContext(context.Context) }); ok {
			contextual.withContext(ctx.(context.Context))
		}
	}
	return job.addJob(c)
}

func bindCronContext(crontab *cron.Cron, ctx context.Context) {
	cronContexts.Store(crontab, ctx)
}

func unbindCronContext(crontab *cron.Cron) {
	cronContexts.Delete(crontab)
}

func (h *HttpJob) addJob(c *cron.Cron) (int, error) {
	id, err := c.AddJob(h.CronExpression, h)
	if err != nil {
		fmt.Println(time.Now().Format(timeFormat), " [ERROR] JobCore AddJob error", err)
		return 0, err
	}
	EntryId := int(id)
	return EntryId, nil
}

func (e *ExecJob) addJob(c *cron.Cron) (int, error) {
	id, err := c.AddJob(e.CronExpression, e)
	if err != nil {
		fmt.Println(time.Now().Format(timeFormat), " [ERROR] JobCore AddJob error", err)
		return 0, err
	}
	EntryId := int(id)
	return EntryId, nil
}

// Remove 移除任务
func Remove(c *cron.Cron, entryID int) chan bool {
	ch := make(chan bool)
	go func() {
		c.Remove(cron.EntryID(entryID))
		fmt.Println(time.Now().Format(timeFormat), " [INFO] JobCore Remove success ,info entryID :", entryID)
		ch <- true
	}()
	return ch
}

// 任务停止
//func Stop() chan bool {
//	ch := make(chan bool)
//	go func() {
//		global.GADMCron.Stop()
//		ch <- true
//	}()
//	return ch
//}
