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
	StartIndex    int // inclusive
	EndIndex int // exclusive
	Entries    []Entry
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
	StartIndex  int
	EndIndex int
	Entries    []Entry
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
	StartIndex    int // inclusive
	EndIndex int // exclusive
}

type QueryResponse struct {
	SenderID int
	View     int
	Entries    []Entry
}

type Query struct {
	Request  QueryRequest
	Response QueryResponse
}

type CoordinateRequest struct {
	SenderID int
	View     int
	StartIndex    int //inclusive
	EndIndex	int //exclusive
	Prepare  bool
	Entries   []Entry
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
	StartIndex int
	EndIndex int
	Entries []Entry
	Sync bool
}
