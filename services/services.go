package services

type Service interface {
	Process(req string) string
  MarshalJSON() ([]byte, error)
  UnmarshalJSON(snap []byte) error
}

func StartService(config string) Service {
  var serv Service
  switch config {
  case "kv-store":
    serv = newStore()
  case "dummy":
    serv = newDummy()
  }
  return serv
}
