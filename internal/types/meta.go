package types

type ServerMeta struct {
	Name    string
	Version string
}

func (m ServerMeta) UA() string {
	if len(m.Version) == 0 {
		return m.Name
	}
	return m.Name + "@" + m.Version
}

type RequestMeta struct {
	ID     string
	Method string
	Route  string
}

type OperationMeta struct {
	ServerMeta
	RequestMeta
}

func (m OperationMeta) UA() string {
	id := m.ID
	if len(id) == 0 {
		id = "unknown"
	}
	return m.ServerMeta.UA() + "(" + id + ")"
}

type OperationMetaProvider interface {
	GetOperation(id string) (OperationMeta, bool)
}
