package types
import (
	"context"
	agent "agentops/common"
)
type AgentService interface{
ExecuteCommands(context.Context,*agent.ExecuteCommandsRequest) error
CreateFiles(context.Context,*agent.CreateFileRequest) error
ViewFiles(context.Context,*agent.ViewFileRequest) (error, string)
}