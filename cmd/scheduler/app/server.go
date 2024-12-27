/* ==================================================================
* Copyright (c) 2024, HIT Authors
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY Wanyou Wang,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: Wanyou Wang
 */

package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/component-base/version"
	"hit.edu/framework/pkg/scheduler"
	"hit.edu/framework/pkg/server"
)

func init() {

}

const SchedulerName = "scheduler"

func NewSchedulerCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:  "scheduler",
		Long: `调度器`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(cmd)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
	}

	return cmd
}

func runCommand(cmd *cobra.Command) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		stopCh := server.SetupSignalHandler()
		<-stopCh
		cancel()
	}()

	sched, err := Setup(ctx)
	if err != nil {
		return err
	}
	return Run(ctx, sched)
}

func Run(ctx context.Context, sched *scheduler.Scheduler) error {
	logs.Init(SchedulerName)
	logs.Info("Starting Scheduler\t", "version\t", version.Get())

	sched.Run(ctx)
	logs.Error("Failed to start scheduler")
	return fmt.Errorf("")
}

func Setup(ctx context.Context) (*scheduler.Scheduler, error) {
	sched, err := scheduler.New(ctx)
	return sched, err
}
