// package main

// import (
//     // "regexp"
//     "fmt"
//     "os/exec"
// )

// func main() {
//     input := []string{
//         "Action{789}.Runtime{2024}",          // 缺省Task和Group
//         "Task{A1}.Action{789}.Runtime{2024}", // 缺省Group
//         "Group{B2}.Action{789}.Runtime{2024}",// 缺省Task
//         "Task{A1}.Group{B2}.Action{789}.Runtime{2024}", // 完整包含
//     }

//     re := regexp.MustCompile(`^(Task\{([^}]+)\})?(\.Group\{([^}]+)\})?\.Action\{([^}]+)\}\.Runtime\{([^}]+)\}$`)

//     for _, s := range input {
//         matches := re.FindStringSubmatch(s)
// 		// if len(matches) == 
//         // if len(matches) >= 6 {
//         //     fmt.Printf("原始文本: %s\n", s)
//         //     fmt.Printf("Task ID: %s\n", matches[2])
//         //     fmt.Printf("Group ID: %s\n", matches[4])
//         //     fmt.Printf("Action ID: %s\n", matches[5])
//         //     fmt.Printf("Runtime ID: %s\n\n", matches[6])
//         // }
// 		fmt.Printf("%v", matches)
//     }
// }

// func main() {
//     cmd := "python"
//     args := "D:\\postgraduate\\project\\heongtong_yolo\\predict.py"
//     CMD := exec.Command(cmd, args)

//     if err := CMD.Start(); err != nil {
// 	    fmt.Errorf("failed to start command: %v", err)
// 	}
//     fmt.Printf("???%v", CMD.Process.Pid)
//     if err := CMD.Wait(); err != nil {
// 		// 检查 stopSignal 通道是否被关闭，判断进程是否是外部停止的
// 		fmt.Printf("command killed externally by stopCMD")
//     }
// }

package main

import (
	"fmt"
	"regexp"
)

func main() {
	// 定义正则表达式
	re := regexp.MustCompile(`(?:Task\{(\d+)\}\.)?(?:Group\{(\d+)\}\.)?Action\{(\d+)\}\.Runtime\{(\d+)\}`)

	// 测试字符串
	testCases := []string{
		"Task{1}.Group{2}.Action{3}.Runtime{4}",
		"Group{2}.Action{3}.Runtime{4}",
		"Action{3}.Runtime{4}",
        "Task{1}.Action{3}.Runtime{4}",
	}

	// 匹配并提取 ID
	for _, testCase := range testCases {
		matches := re.FindStringSubmatch(testCase)
		if matches != nil {
			fmt.Printf("匹配的字符串: %s\n", testCase)
			fmt.Printf("提取的 ID: TaskID=%s, GroupID=%s, ActionID=%s, RuntimeID=%s\n",
				matches[1], matches[2], matches[3], matches[4])
		} else {
			fmt.Printf("未匹配到: %s\n", testCase)
		}
	}
}