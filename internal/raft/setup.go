package raft

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

func SetupRaft(nodeID string,bindAddr string,baseDir string, fsm raft.FSM) (*raft.Raft, error){
	config:=raft.DefaultConfig();
	config.LocalID=raft.ServerID(nodeID)

	if err:=os.Mkdir(baseDir,0700); err!=nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	addr,err:=net.ResolveTCPAddr("tcp",bindAddr)
	if err!=nil{
		return nil,err
	}
	transport,err:=raft.NewTCPTransport(bindAddr,addr,3,10*time.Second,os.Stderr)
	if err!=nil{
		return nil,err
	}

	boltDBPath:=filepath.Join(baseDir,"raft.db")
	boltStore,err:=raftboltdb.NewBoltStore(boltDBPath)
	if err!=nil{
		return nil,fmt.Errorf("failed to create BoltDB store: %w", err)
	}
	snapshotStore,err:=raft.NewFileSnapshotStore(baseDir,2,os.Stderr)
	if err!=nil{
		return nil,fmt.Errorf("failed to create snapshot store: %w", err)
	}

	raftNode,err:=raft.NewRaft(config,fsm,boltStore,boltStore,snapshotStore,transport)
	if err!=nil{
		return nil,fmt.Errorf("failed to create Raft node: %w", err)
	}

	hasState,err:=raft.HasExistingState(boltStore,boltStore,snapshotStore)
	if err!=nil{
		return nil, fmt.Errorf("failed to check existing Raft state: %w", err)
	}

	if !hasState{
		slog.Info("Bootstrapping new Raft cluster", slog.String("node_id", nodeID))
		configuration:=raft.Configuration{
			Servers: []raft.Server{
				{
					ID: config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		future:=raftNode.BootstrapCluster(configuration)
		if err:=future.Error(); err!=nil{
			return nil, fmt.Errorf("failed to bootstrap Raft cluster: %w", err)
		}
	} else {
		slog.Info("Recovering existing Raft state", slog.String("node_id", nodeID))
	}

	return raftNode,nil
}