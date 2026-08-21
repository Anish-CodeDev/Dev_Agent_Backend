package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	agent "agentops/common"
)
type httpServer struct {
	addr string
}

type requestBody struct{
	Files    []string `json:"files"`
	Contents []string `json:"contents"`
	AppName  string   `json:"app_name"`
}
type commandBody struct{
	Cmd []string `json:"commands"`
}

type folder_path struct{
	Path string `json:"file_path"`
}
func NewHttpServer(addr string) *httpServer {
	return &httpServer{addr: addr}
}

func (s *httpServer) Run() error {
	router := http.NewServeMux()

	conn := NewGRPCClient(":9000")
	defer conn.Close()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req requestBody
		defer r.Body.Close()
		decoder:=json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}
		if len(req.Files) == 0 || len(req.Files) != len(req.Contents) {
			http.Error(w, "files and contents must be non-empty and equal length", http.StatusBadRequest)
			return
		}
		if req.AppName == "" {
			http.Error(w, "app_name is required", http.StatusBadRequest)
			return
		}

		c := agent.NewManageAgentOpsClient(conn)
		ctx, cancel := context.WithTimeout(r.Context(), time.Second * 60 * 2)
		defer cancel()
		res,err := c.CreateFiles(ctx,&agent.CreateFileRequest{
			Files:req.Files,
			Contents: req.Contents,
			AppName: req.AppName,
		})
		if err!=nil{
			http.Error(w, "Request failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(res)
		if err!=nil{
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

		
	})

	router.HandleFunc("/commands", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var cmd commandBody
		defer r.Body.Close()
		decoder:=json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&cmd); err != nil {
			http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}
		if len(cmd.Cmd) == 0{
			http.Error(w, "files and contents must be non-empty and equal length", http.StatusBadRequest)
			return
		}
		

		c := agent.NewManageAgentOpsClient(conn)
		ctx, cancel := context.WithTimeout(r.Context(), time.Second * 60 * 2)
		defer cancel()
		res,err := c.ExecuteCommands(ctx,&agent.ExecuteCommandsRequest{
			Cmds:  cmd.Cmd,
		})
		if err!=nil{
			http.Error(w, "Request failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(res)
		if err!=nil{
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return 
		}

		
	})

	router.HandleFunc("/view_code", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var folder folder_path
		defer r.Body.Close()
		decoder:=json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&folder); err != nil {
			http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}
		if folder.Path == ""{
			http.Error(w, "Enter a valid path", http.StatusBadRequest)
			return
		}
		

		c := agent.NewManageAgentOpsClient(conn)
		ctx, cancel := context.WithTimeout(r.Context(), time.Second * 60 * 2)
		defer cancel()
		res,err := c.ViewFile(ctx,&agent.ViewFileRequest{
			Path: folder.Path,
		})
		if err!=nil{
			http.Error(w, "Request failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(res)
		if err!=nil{
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return 
		}

		
	})

	log.Println("Starting server on", s.addr)

	return http.ListenAndServe(s.addr, router)
}