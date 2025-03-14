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

type ConditionVal int32

const (
	TrueVal           ConditionVal = 1
	FalseVal          ConditionVal = 2
	NotReadyToCompute ConditionVal = 3
)

type Condition interface {
	GetVal() (res ConditionVal)
	CheckReady() bool
}

type TrueCondition struct{}

func (c TrueCondition) getVal() (res ConditionVal) {
	if !c.checkReady() {
		return NotReadyToCompute
	}
	return TrueVal
}

func (c TrueCondition) checkReady() bool {
	return true
}

type FalseCondition struct{}

func (c FalseCondition) checkReady() bool {
	return true
}

func (c FalseCondition) getVal() (res ConditionVal) {
	return FalseVal
}

// TODO: waiting to complete...
type EqualCondition struct {
}

func (c EqualCondition) checkReady() bool {
	return true
}

func (c EqualCondition) getVal() (res ConditionVal) {
	return FalseVal
}
