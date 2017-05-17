package service

// Service is a configurable process
// that must remain running
type Service struct {
	ID          string   `json:"id"`
	Args        []string `json:"args"`
	Environment []*Env   `json:"environment"`
	Files       []*File  `json:"files"`
}

func (s Service) Env(name string) *Env {
	for _, env := range s.Environment {
		if env.Name == name {
			return env
		}
	}
	return nil
}
