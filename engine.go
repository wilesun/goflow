package goflow

import (
	"os"
	"time"

	"github.com/lunny/log"
)

type Engine struct {
	processService *ProcessService //流程定义业务类
	orderService   *OrderService   //流程实例业务类
	taskService    *TaskService    //任务业务类
	queryService   *QueryService   //查询业务类
	managerService *ManagerService //管理业务类
}

func (p *Engine) StartInstanceById(id string, operator string, args map[string]interface{}) *Order {
	process := new(Process)
	process.GetProcessById(id)
	return p.StartProcess(process, operator, args)
}

func (p *Engine) StartInstanceByName(name string, version int, operator string, args map[string]interface{}) *Order {
	process := new(Process)
	process.GetProcessByVersion(name, version)
	return p.StartProcess(process, operator, args)
}

func (p *Engine) StartProcess(process *Process, operator string, args map[string]interface{}) *Order {
	execution := p.ExecuteByProcess(process, operator, args)
	if process.Model != nil {
		start := process.Model.GetStart()
		start.Execute(execution)
	}
	return execution.Order
}

func (p *Engine) ExecuteByProcess(process *Process, operator string, args map[string]interface{}) *Execution {
	order := p.orderService.CreateOrder(process, operator)
	execution := &Execution{
		Engine:   p,
		Process:  process,
		Order:    order,
		Operator: operator,
	}
	return execution
}

func (p *Engine) ExecuteByTaskId(id string, operator string, args map[string]interface{}) *Execution {
	//todo
	task := p.taskService.Complete(id, operator, args)

	order := &Order{}
	order.GetOrderById(task.OrderId)
	order.LastUpdator = operator
	order.LastUpdateTime = time.Now()
	if task.TaskType != TT_MAJOR {
		return nil
	} else {
		process := new(Process)
		process.GetProcessById(order.ProcessId)

		execution := &Execution{
			Engine:   p,
			Process:  process,
			Order:    order,
			Operator: operator,
			Task:     task,
		}
		return execution
	}
}

func (p *Engine) ExecuteAndJumpTask(id string, operator string, args map[string]interface{}, nodeName string) []*Task {
	execution := p.ExecuteByTaskId(id, operator, args)
	if execution != nil {
		model := execution.Process.Model
		if nodeName == "" {
			task := p.taskService.RejectTask(model, execution.Task)
			execution.AddTask(task)
		} else {
			nodeModel := model.GetNode(nodeName)
			tm := &TransitionModel{
				Target:  nodeModel,
				Enabled: true,
			}
			tm.Execute(execution)
		}

		return execution.Tasks
	}
	return []*Task{}
}

func NewEngine() *Engine {
	return nil
}

func init() {
	f, _ := os.Create("goflow.log")
	log.Std.SetOutput(f)
}
