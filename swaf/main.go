package main

import ("encoding/json"
"fmt"
"log"
"mime"
"net/http"
"os"
"strconv"
"time"
"swaf/internal/taskstore"
)
type taskServer struct{
	store *taskstore.TaskStore
}
func NewTaskServer() *taskServer{
	store:= taskstore.New()
	return &taskServer{store:store}
}
func (ts *taskServer) createTaskHandler (w http.ResponseWriter ,req *http.Request){
	log.Printf("handling task creat at %s \n",req.URL.Path)
  
	type RequestTask struct{
		Text string `json:"text"`
		Tags []string `json:"tags"`
		Due time.Time `json:"due"`


	}
	type ResponseId struct {
		Id int `json:"id"`
	}
	contentType := req.Header.Get("Content-Type")
	mediatype,_,err := mime.ParseMediaType(contentType)
	if err!=nil{
		http.Error(w,err.Error(),http.StatusBadRequest)
		return 
	}
	if mediatype !="application/json"{
		http.Error(w,"expect application/json Content-Type",http.StatusUnsupportedMediaType)

	}
	dec:= json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var rt RequestTask
	if err:= dec.Decode(&rt);err!=nil{
		http.Error(w,err.Error(),http.StatusBadRequest)
		return 
	}
  id:= ts.store.CreateTask(rt.Text,rt.Tags , rt.Due)
  js,err := json.Marshal(ResponseId{Id:id});
  if err!=nil{
	http.Error(w,err.Error(),http.StatusInternalServerError)
	return
  }
  w.Header().Set("Content-Type","application/json")
  w.Write(js);

	

}
func (ts *taskServer) getTaskHandler(w http.ResponseWriter, req *http.Request){
	log.Printf("handling get task at %s\n",req.URL.Path)

	id,err := strconv.Atoi(req.PathValue("id"));
	if err!= nil{
		http.Error(w,"invalid id",http.StatusBadRequest)
		return 
	}
	task,err := ts.store.GetTask(id);
	if err !=nil{
		http.Error(w,err.Error(),http.StatusNotFound)
		return 
	}
	js,err:= json.Marshal(task)
	if err != nil{
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return 
	}
	w.Header().Set("Content-Type","application/json")
	w.Write(js)
}
func(ts *taskServer) deleteTaskHandler(w http.ResponseWriter, req *http.Request){
	log.Printf("handling delete task at %s\n",req.URL.Path)

	id,err := strconv.Atoi(req.PathValue("id"))
	if err!= nil{
		http.Error(w,"invalid id",http.StatusBadRequest)
		return 
	}
	err = ts.store.DeleteTask(id)
	if err!=nil{
		http.Error(w,err.Error(),http.StatusNotFound)

	}
}
func (ts *taskServer) deleteAllTasksHandler(w http.ResponseWriter,req *http.Request){
	log.Printf("handling delete all task at %s\n", req.URL.Path)
	ts.store.DeleteAllTasks()

}
func (ts *taskServer)tagHandler(w http.ResponseWriter , req *http.Request){
	log.Printf("handling tasks by tag at %s\n",req.URL.Path)

	tag:= req.PathValue("tag")
	tasks:= ts.store.GetTasksByTag(tag)
	js,err := json.Marshal(tasks)
     if err!=nil{
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return 
	 }
	 w.Header().Set("Content-Type","application/json")
	 w.Write(js)
}
func (ts *taskServer)dueHandler(w http.ResponseWriter, req *http.Request){
	log.Printf("handling tasks by dye at %s\n",req.URL.Path)
	badRequestError:=func(){
		http.Error(w,fmt.Sprintf("expect /due/<year>/<month>/<day>,got %v",req.URL.Path),http.StatusBadRequest)

	}
	year,errYear:= strconv.Atoi(req.PathValue("year"))
	month,errMonth := strconv.Atoi(req.PathValue("month"))
	day,errDay := strconv.Atoi(req.PathValue("day"))
	if errYear !=nil || errMonth!=nil || errDay!=nil || month<int(time.January) || month > int(time.December){
		badRequestError()
		return
	}
	tasks:= ts.store.GetTasksByDueDate(year,time.Month(month),day)
	js,err:= json.Marshal(tasks)
	if err !=nil{
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return 
	}
	w.Header().Set("Content-Type","application/json")
	w.Write(js)
}
func main(){
	mux:= http.NewServeMux()
	server:= NewTaskServer()
	mux.HandleFunc("POST /task/", server.createTaskHandler)
	mux.HandleFunc("GET /task/", server.getAllTasksHandler)
	mux.HandleFunc("DELETE /task/", server.deleteAllTasksHandler)
	mux.HandleFunc("GET /task/{id}/", server.getTaskHandler)
	mux.HandleFunc("DELETE /task/{id}/", server.deleteTaskHandler)
	mux.HandleFunc("GET /tag/{tag}/", server.tagHandler)
	mux.HandleFunc("GET /due/{year}/{month}/{day}/", server.dueHandler)

	log.Fatal(http.ListenAndServe("localhost:"+os.Getenv("SERVERPORT"), mux))
}