package main 
import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"strconv"
)


func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := getTpl(w,"index.tpl")
	if tmpl != nil {
		data := new(HtmlData)
		data.Title = "GoHAProxy"
		for _,proxy := range proxyServer.ServerList {
			data.ProxyStatus = append(data.ProxyStatus, proxy.srvProxy)
	    }
		tmpl.Execute(w, data)	
	}
}


func ApiHandler(w http.ResponseWriter, r *http.Request) {
	/*vars := mux.Vars(r)
	now := time.Now().Unix()
	data := getModuleRecord(vars["method"], now - 86400, now)
	json, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(500)
		log.Println("Error generating json", err)
		fmt.Fprintln(w, "Could not generate JSON")
	}		
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(json))*/		
}


func liveApiHandler(w http.ResponseWriter, r *http.Request) {
	/*vars := mux.Vars(r)
	data := getModuleLastRecord(vars["siteId"], vars["method"])
	json, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(500)
		log.Println("Error generating json", err)
		fmt.Fprintln(w, "Could not generate JSON")
	}		
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(json))*/
}


func SystemApiHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)	
	result := make(map[string]string)
	result["cmd"] = vars["cmd"]	
	
	switch vars["cmd"] {
		case "reload":
 			
	}
	
	
	json, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(500)
		log.Println("Error generating json", err)
		fmt.Fprintln(w, "Could not generate JSON")
	}		
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(json))
}



func strToInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
			   // handle error
	  fmt.Println(err)
	  return 0
	}		  
	return i  
}


func getTpl(w http.ResponseWriter, tpl string) (*template.Template, error) {
	tmpl, err := template.ParseFiles("./tpl/" + tpl)
	if err != nil {
		log.Println("Could not parse template", err)
		fmt.Fprintln(w, "Problem parsing template", err)		
	}
	return tmpl, err
}


func Monitor() {		
   	httpServer()   	
}


func httpServer() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	//System
	r.HandleFunc("/api/system/{cmd}", SystemApiHandler)
	
	http.Handle("/", r)
	fmt.Println("Monitor server started.")
	http.ListenAndServe(":8080", nil)
}
