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
package cli

import "github.com/spf13/cobra"

func Run(cmd *cobra.Command) int {
	if _, err := run(cmd); err != nil {
		return 1
	}
	return 0
}

func run(cmd *cobra.Command) (logsInitialized bool, err error) {
	err = cmd.Execute()
	return
}
