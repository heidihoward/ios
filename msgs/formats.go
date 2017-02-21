package msgs

// MESSAGE FORMATS

type ClientRequest struct {
	ClientID  int
	RequestID int
	Replicate bool
	ForceViewChange bool
	Request   string
}

type ClientResponse struct {
	ClientID  int
	RequestID int
	Success bool
	Response  string
}

type Entry struct {
	View      int
	Committed bool
	Requests  []ClientRequest
}

type PrepareRequest struct {
	SenderID int
	View     int
	Index    int
	Entry    Entry
}

type PrepareResponse struct {
	SenderID int
	Success  bool
}

type Prepare struct {
	Request  PrepareRequest
	Response PrepareResponse
}

type CommitRequest struct {
	SenderID int
	View     int // TODO: remove view
	Index    int
	Entry    Entry
}

type CommitResponse struct {
	SenderID    int
	Success     bool
	CommitIndex int
}

type Commit struct {
	Request  CommitRequest
	Response CommitResponse
}

type NewViewRequest struct {
	SenderID int
	View     int
}

type NewViewResponse struct {
	SenderID int
	View     int
	Index    int
}

type NewView struct {
	Request  NewViewRequest
	Response NewViewResponse
}

type QueryRequest struct {
	SenderID int
	View     int
	Index    int
}

type QueryResponse struct {
	SenderID int
	View     int
	Present  bool
	Entry    Entry
}

type Query struct {
	Request  QueryRequest
	Response QueryResponse
}

type CoordinateRequest struct {
	SenderID int
	View     int
	Index    int
	Prepare  bool
	Entry    Entry
}

type CoordinateResponse struct {
	SenderID int
	Success  bool
}

type Coordinate struct {
	Request  CoordinateRequest
	Response CoordinateResponse
}

type LogUpdate struct {
	Index int
	Entry Entry
}
