package core

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/core/filter"
	"quorumengineering/quorum-report/core/monitor"
	"quorumengineering/quorum-report/core/rpc"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

// Backend wraps MonitorService and QuorumClient, controls the start/stop of the reporting tool.
type Backend struct {
	lastPersisted uint64
	monitor       *monitor.MonitorService
	filter        *filter.FilterService
	rpc           *rpc.RPCService
}

func New(config types.ReportInputStruct) (*Backend, error) {
	quorumClient, err := client.NewQuorumClient(config.Reporting.WSUrl, config.Reporting.GraphQLUrl)
	if err != nil {
		return nil, err
	}
	db := database.NewMemoryDB()
	lastPersisted := db.GetLastPersistedBlockNumber()
	if len(config.Reporting.Addresses) > 0 {
		err = db.AddAddresses(config.Reporting.Addresses)
		if err != nil {
			return nil, err
		}
	}
	return &Backend{
		lastPersisted: lastPersisted,
		monitor:       monitor.NewMonitorService(db, quorumClient),
		filter:        filter.NewFilterService(db),
		rpc:           rpc.NewRPCService(db, config.Reporting.RPCAddr, config.Reporting.RPCVHosts, config.Reporting.RPCCorsList),
	}, nil
}

func (b *Backend) Start() {
	// Start monitor service.
	go b.monitor.Start()
	// Start filter service.
	go b.filter.Start(b.lastPersisted)
	// Start local RPC service.
	go b.rpc.Start()

	// cleaning...
	defer func() {
		b.rpc.Stop()
		b.filter.Stop()
	}()

	// Keep process alive before killed.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("Process stopped by SIGINT or SIGTERM.")
}