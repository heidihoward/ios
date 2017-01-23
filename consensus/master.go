package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"time"
)

var noop = msgs.ClientRequest{-1, -1, true, "noop"}

// RunMaster implements the Master mode
func RunMaster(view int, commit_index int, inital bool, io *msgs.Io, config Config) {
	// setup
	glog.Info("Starting up master in view ", view)
	majority := Majority(config.N)

	// determine next safe index
	index := -1
	if !inital {
		// dispatch new view requests
		req := msgs.NewViewRequest{config.ID, view}
		(*io).OutgoingBroadcast.Requests.NewView <- req

		// collect responses
		glog.Info("Waiting for ", majority, " new view responses")
		min_index := commit_index
		// TODO: FEATURE add option to wait longer

		for i := 0; i < majority; {
			msg := <-(*io).Incoming.Responses.NewView
			// check msg replies to the msg we just sent
			if msg.Request == req {
				res := msg.Response
				glog.Info("Received ", res)
				if res.Index > index {
					index = res.Index
				} else if res.Index < min_index {
					min_index = res.Index
				}
				i++
				// TODO: BUG need to check view
				glog.Info("Successful new view received, waiting for ", majority-i, " more")
			}

		}
		glog.Info("Index is ", index)

		// recover entries
		for curr_index := commit_index + 1; curr_index <= index; curr_index++ {
			RunRecoveryCoordinator(view, curr_index, io, config)
		}

	}

	if config.BatchInterval == 0 {
		glog.Info("Ready to handle requests. No batching enabled")
		// handle client requests (1 at a time)
		for {

			// wait for request
			req := <-io.IncomingRequests
			glog.Info("Request received: ", req)

			// if possible, handle request without replication
			if !req.Replicate {
				io.OutgoingRequests <- req
				glog.Info("Request handled with replication: ", req)
			} else {
				index++
				glog.Info("Request assigned index: ", index)
				ok := RunCoordinator(view, index, []msgs.ClientRequest{req}, io, config, true)
				if !ok {
					break
				}
				glog.Info("Finished replicating request: ", req)
			}

		}
	} else {
		glog.Info("Ready to handle request. Batch every ", config.BatchInterval, " microseconds")
		for {
			// setup for holding requests
			reqs := make([]msgs.ClientRequest, config.MaxBatch)
			reqs_num := 0

			// start collecting requests
			timeout := make(chan bool, 1)
			go func() {
				<-time.After(time.Microsecond * time.Duration(config.BatchInterval))
				timeout <- true
			}()

			exit:= false
		    for ;exit==false; {
		    	select {
		    	case req := <-io.IncomingRequests:
					if !req.Replicate {
						io.OutgoingRequests <- req
						glog.Info("Request handled without replication: ", req)
					} else {
						reqs[reqs_num] = req
						glog.Info("Request ",reqs_num, " is : ", req)
						reqs_num = reqs_num + 1
						if reqs_num==config.MaxBatch {
							exit=true
							break
						}
					}
				case <-timeout:
					exit=true
					break
				}
		    }


		    // assign requests to coordinators
		    if reqs_num>0 {
		    	glog.Info("Starting to replicate ",reqs_num, " requests")
				reqs_small := reqs[:reqs_num]
				index++
				ok := RunCoordinator(view, index, reqs_small, io, config, true)
				if !ok {
					break
				}
				glog.Info("Finished replicating requests: ", reqs_small)
			}
		}

	}


	glog.Warning("Master stepping down")

}
