//in memory data store map[int]Task
package taskstore

import (
	"fmt"
	"sync"
	"time"
)
type Task struct {
	Id int          `json:"id"`
	Text string     `json:"test"`
	Tags []string    `json:"tags"`
	Due  time.Time   `json:"due"`
}
//Taskstore is inmemory database of task taskstore method are 
//safe to call concurrently
type TaskStore struct {
	sync.Mutex
	tasks map[int]Task
	nextId int

}
func New() *TaskStore{
	//when we are using contructor it
	//noting but we are just making a
	//method in which we are 
	//initializintg that object 
	//and then assinging values to it 
	ts:= &TaskStore{}
	ts.tasks= make(map[int]Task)
	ts.nextId=0;
	return ts;
}
//create new task in store
func (ts *TaskStore) CreateTask(text string, tags []string,due time.Time) int{
	ts.Lock()
	defer ts.Unlock()
	task:= Task{
		Id : ts.nextId,
		Text: text,
		Due:due,
	}
	task.Tags = make([]string,len(tags));
	copy(task.Tags,tags)//copy one string to another
	ts.tasks[ts.nextId]=task
	ts.nextId++
	return task.Id

}
//get taks uses id to retrive task

func(ts *TaskStore)GetTask(id int)(Task,error){
	ts.Lock()
	defer ts.Unlock()
	t,ok := ts.tasks[id]
	if(ok){
		return t,nil

	}else{
		return Task{},fmt.Errorf("task with id=%d not found",id)
	}
}
//delete task with in the memory
func (ts *TaskStore) DeleteTask(id int)error{
	ts.Lock()
	defer ts.Unlock()
	if _,ok := ts.tasks[id];!ok {
		return fmt.Errorf("task with id=%d not found",id)
	}
	delete(ts.tasks,id)
	return nil
}
//delete all tasks 
func (ts *TaskStore)DeleteAllTasks()error{
	ts.Lock()
	defer ts.Unlock()
	ts.tasks= make(map[int]Task)
	return nil
}
//get all task
func(ts *TaskStore) GetAllTask()[]Task{
	ts.Lock()
	defer ts.Unlock()
	allTask := make([]Task,0,len(ts.tasks))
	for _,task := range ts.tasks{
		allTask = append(allTask,task)
	}
	return allTask
}
//get task by tag
func(ts *TaskStore)GetTasksByTag(tag string)[]Task{
	ts.Lock()
	defer ts.Unlock()

	var tasks []Task

	taskloop: 
	          for _,task := range ts.tasks {
				for _,taskTag := range task.Tags{
					if taskTag == tag{
						tasks =append(tasks,task)
						continue taskloop//for moving to next iteration to next upper loop

					}
				}
			  }
			  return tasks
}
//get task by due data
func(ts *TaskStore)GetTasksByDueDate(year int ,month time.Month , day int)[]Task{
	ts.Lock()
	defer ts.Unlock()

	var tasks []Task
	for _,task := range ts.tasks{
		y,m,d := task.Due.Date()
		if y== year && m== month && d==day{
			tasks= append(tasks,task)
		}
	}
	return tasks
}
