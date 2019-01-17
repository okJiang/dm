// Copyright 2018 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package master

import (
	"github.com/pingcap/dm/dm/ctl/common"
	"github.com/pingcap/dm/dm/pb"
	"golang.org/x/net/context"
)

// operateTask does operation on task
func operateTask(op pb.TaskOp, name string, workers []string) (*pb.OperateTaskResponse, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cli := common.MasterClient()
	return cli.OperateTask(ctx, &pb.OperateTaskRequest{
		Op:      op,
		Name:    name,
		Workers: workers,
	})
}
