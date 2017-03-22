package msgs

// MESSAGE FORMATS

// ClientRequest desribes a request.
type ClientRequest struct {
	ClientID        int
	RequestID       int
	ForceViewChange bool
	Request         string
}

// ClientResponse desribes a request response.
type ClientResponse struct {
	ClientID  int
	RequestID int
	Success   bool
	Response  string
}

// Client wraps a ClientRequest and ClientResponse.
type Client struct {
	Request  ClientRequest
	Response ClientResponse
}

// Entry describes a item stored in the replicated log.
type Entry struct {
	View      int
	Committed bool
	Requests  []ClientRequest
}

// PrepareRequest describes a Prepare messages.
type PrepareRequest struct {
	SenderID   int
	View       int
	StartIndex int // inclusive
	EndIndex   int // exclusive
	Entries    []Entry
}

// PrepareResponse describes a Prepare response messages.
type PrepareResponse struct {
	SenderID int
	Success  bool
}

// Prepare wraps a PrepareRequest and PrepareResponse.
type Prepare struct {
	Request  PrepareRequest
	Response PrepareResponse
}

// CommitRequest describes a Commit messages.
type CommitRequest struct {
	SenderID   int
	StartIndex int
	EndIndex   int
	Entries    []Entry
}

// CommitResponse describes a Commit response messages.
type CommitResponse struct {
	SenderID    int
	Success     bool
	CommitIndex int
}

// Commit wraps a CommitRequest and CommitResponse.
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
	SenderID   int
	View       int
	StartIndex int // inclusive
	EndIndex   int // exclusive
}

type QueryResponse struct {
	SenderID int
	View     int
	Entries  []Entry
}

type Query struct {
	Request  QueryRequest
	Response QueryResponse
}

type CoordinateRequest struct {
	SenderID   int
	View       int
	StartIndex int //inclusive
	EndIndex   int //exclusive
	Prepare    bool
	Entries    []Entry
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
	EndIndex   int
	Entries    []Entry
}
