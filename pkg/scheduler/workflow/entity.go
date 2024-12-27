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
package workflow

import condition "hit.edu/framework/pkg/scheduler/workflow/condition"

type ActionType int32
type ResourceType int32

const (
	CMDAction ActionType = 1
	JSAction  ActionType = 2

	CPU     ResourceType = 1
	Memory  ResourceType = 2
	GPU     ResourceType = 3
	Network ResourceType = 4
)

type Group struct {
	GroupId        string
	actions        []Action
	GroupCondition condition.Condition
	Spec           GroupSpec
}

type GroupSpec struct {
	SchedulerName string
	NodeName      string
}

type Action struct {
	actionCondition condition.Condition
	actionType      ActionType
	actUrl          string
	actCMD          string
	resourceQuota   []ResourceQuota
}

type ResourceQuota struct {
	resourceType   ResourceType
	resourceDemand string
	meta           ResourceMata
}

type ResourceMata struct {
	valueType string
	lowBound  bool
	upBound   bool
}
