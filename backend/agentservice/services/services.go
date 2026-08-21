package services

import (
	agent "agentops/common"
	"context"
	"fmt"
	"os"
	"os/exec"
	"github.com/mattn/go-shellwords"
)

func runCommand(command string) (string, error) {
	args, err := shellwords.Parse(command)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

type AgentService struct {
}

func NewAgentService() *AgentService {
	return &AgentService{}
}
func (s *AgentService) ExecuteCommands(ctx context.Context, in *agent.ExecuteCommandsRequest) error {

	if err:= os.Mkdir("/data",0755);err!=nil{
		fmt.Println("Directory already exists")
	}
	res:=fmt.Sprintf("/data/%v","test")
	if err:=os.Mkdir(res,0755);err!=nil{
		fmt.Println("Directory already exists")
	}
	for _,cmd:= range in.Cmds{
		cmd += "\n"
		data:= []byte(cmd)
		err:=os.WriteFile(res + "/commands.txt",data,0755)
		if err!=nil{
			return err
		}

	}
	return nil
}

func (s *AgentService) CreateFiles(ctx context.Context, in *agent.CreateFileRequest) error {
	file_names := []string{}
	for _, val := range in.Files {
		file_names = append(file_names, val)
	}
	if err := os.Mkdir("/data", 0755); err != nil {
		fmt.Println("Error creating /data:", err)
	}
	res := fmt.Sprintf("/data/%v/", in.AppName)
	if err := os.Mkdir(res, 0755); err != nil {
		fmt.Println("Directory already exists")
	}
	for i, val := range in.Contents {
		data := []byte(val)
		file_name := fmt.Sprintf("/data/%v/%v", in.AppName, file_names[i])
		err := os.WriteFile(file_name, data, 0644)
		if err != nil {
			return err
		}
	}
	
	return nil
}


func(s *AgentService) ViewFiles(ctx context.Context, in *agent.ViewFileRequest)(error,string){
fmt.Println("User wants to access the folder located at ",in.Path);
data,err := os.ReadFile(fmt.Sprintf("/data/%s",in.Path))
if err!=nil{
	return nil,"Failed";
}
return nil,string(data)
}