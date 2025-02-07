package linux

import (
	lnx "github.com/faelmori/keepgo/internal/linux"
	"github.com/faelmori/keepgo/runners"
	"github.com/faelmori/keepgo/service"
	"testing"
)

type systemdService struct {
	Name     string
	Config   *service.Config
	i        service.Controller
	platform string
	runner   *runners.Runner
}

func (s *systemdService) Start() error {
	return nil
}

func TestSystemdServiceStatus(t *testing.T) {
	config := &service.Config{Name: "test-service"}
	ctrl := &mockController{}
	//runner := &RunnerImpl{}
	svc, err := lnx.NewSystemdService(ctrl, "linux-systemd", config)
	if err != nil {
		t.Fatalf("Failed to create systemd service: %v", err)
	}

	status, err := svc.Status()
	if err != nil {
		t.Fatalf("Failed to get service status: %v", err)
	}
	if status != service.StatusRunning {
		t.Errorf("Expected status %v, got %v", service.StatusRunning, status)
	}
}

type mockController struct{}

func (m *mockController) Start(s service.Service) error { return nil }
func (m *mockController) Stop(s service.Service) error  { return nil }
func (m *mockController) Run() error                    { return nil }

type RunnerImpl struct{}

func (r *RunnerImpl) RunWithOutput(command string, arguments ...string) (int, string, error) {
	return 0, "", nil
}
