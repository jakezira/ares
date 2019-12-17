package ares

type SApp struct {
	name     string
	nodes    []string
	status   int
	lastNode string
}

var (
	AppStatus_On   = 0
	AppStatus_Fail = 1
	AppStatus_Off  = 2
)

func newSApp(name string) *SApp {
	return &SApp{
		name:   name,
		nodes:  make([]string, 0),
		status: AppStatus_Off,
	}
}

func remove(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func (a *SApp) resetNode(node string) {
	for i, inode := range a.nodes {
		if inode == node {
			a.nodes = remove(a.nodes, i)
			return
		}
	}
}

func (a *SApp) registerNode(node string) {
	a.nodes = append(a.nodes, node)
}

func (a *SApp) setStatus(node string, status int) {
	a.status = status
	a.lastNode = node
}
