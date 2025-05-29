package utils

//
//type SceneWorker interface {
//	CheckScene(runtime *apis.Runtime, action *apis.Action) error
//	CheckObjectScene(spec apis.SceneSpec, status apis.SceneStatus) bool
//	CheckPositionScene(spec apis.SceneSpec, status apis.SceneStatus) bool
//}
//
//// TemporaryMap 用来临时存储一个scene的信息，帮助SceneStatus的状态更新
//type TemporaryMap struct {
//	AttachedScene  string
//	AttachedDevice string
//	AttachedTask   string
//	Location       string
//	Time           apis.Time
//}
//
////// CheckScene 检查scene
////func CheckScene(runtime *apis.Runtime, action *apis.Action) error {
////	// TODO:确认action中scene的key
////	scenes := runtime.Scenes
////	for index, scene := range scenes {
////		switch scene.Type {
////		case apis.ObjectType:
////			if CheckObjectScene(scene, action.Status.Scenes[index]) {
////				return nil
////			} else {
////				return fmt.Errorf("object %s not satisfied\n", scene.SceneID)
////			}
////		case apis.PositionType:
////			if CheckPositionScene(scene, action.Status.Scenes[index]) {
////				return nil
////			} else {
////				return fmt.Errorf("position %s not satisfied\n", scene.SceneID)
////			}
////		}
////	}
////	return nil
////}
//
//// CheckObjectScene 检查object类scene
//func CheckObjectScene(spec apis.SceneSpec, status apis.SceneStatus) bool {
//	flag := true
//	// TODO:上锁还是未上锁？
//	// 检查锁的状态
//	if !status.Lock.Lock {
//		flag = false
//		logs.Warnf("Object %s is not locked", spec.SceneID)
//	}
//	// 关联设备不为空
//	if status.AttachedDevice != "" {
//		flag = false
//		logs.Warnf("Object %s is attached to device %s", spec.SceneID, status.AttachedDevice)
//	}
//	// 关联任务不为空
//	if status.AttachedTask != "" {
//		flag = false
//		logs.Warnf("Object %s is attached to task %s", spec.SceneID, status.AttachedTask)
//	}
//	// 检查期待属性是否填写
//	for key, ep := range spec.ExpectedProperty {
//		if ep.Value == "" {
//			flag = false
//			logs.Warnf("Object %s is missing property %s", spec.SceneID, key)
//		}
//	}
//
//	return flag
//}
//
//// CheckPositionScene 检查position类scene
//func CheckPositionScene(spec apis.SceneSpec, status apis.SceneStatus) bool {
//	flag := true
//	// TODO:上锁还是未上锁？
//	// 检查锁的状态
//	if !status.Lock.Lock {
//		flag = false
//		logs.Warnf("Position %s is not locked", spec.SceneID)
//	}
//	// 关联设备不为空
//	if status.AttachedDevice != "" {
//		flag = false
//		logs.Warnf("Position %s is attached to device %s", spec.SceneID, status.AttachedDevice)
//	}
//	// 关联任务不为空
//	if status.AttachedTask != "" {
//		flag = false
//		logs.Warnf("Position %s is attached to task %s", spec.SceneID, status.AttachedTask)
//	}
//	// 检查期待属性是否填写
//	for key, ep := range spec.ExpectedProperty {
//		if ep.Value == "" {
//			flag = false
//			logs.Warnf("Position %s is missing property %s", spec.SceneID, key)
//		}
//	}
//	return flag
//}
//
////// UpdateSceneStatus 更新Scene状态
////func UpdateSceneStatus(runtime *apis.Runtime, action *apis.Action, deviceNumber int, taskId string) error {
////	scenes := runtime.Scenes
////	/* 需要更新的状态有：关联设备 关联任务 时间 关联scene 属性 */
////	// 只有一个设备
////	temporaryMap := GetSceneStatusUpdateMap(runtime, taskId, action)
////	if deviceNumber == 1 {
////		for index, scene := range scenes {
////			switch scene.Type {
////			// object类
////			case apis.ObjectType:
////				// TODO:SceneId Location
////				status := apis.SceneStatus{
////					UpdateTime:     temporaryMap[scene.SceneID].Time,
////					AttachedTask:   temporaryMap[scene.SceneID].AttachedTask,
////					AttachedDevice: temporaryMap[scene.SceneID].AttachedDevice,
////					Property:       map[string]apis.Property{"location": apis.Property{Name: "location"}},
////					Lock:           apis.Lock{Type: apis.MutexLock, IsLocked: true, Ref: action.Status.Scenes[index].Lock.Ref + 1},
////
////					UpdateMethod:  action.Status.Scenes[index].UpdateMethod,
////					AttachedScene: action.Status.Scenes[index].AttachedScene,
////				}
////				action.Status.Scenes[index] = status
////
////			// position类型
////			case apis.PositionType:
////				// TODO:SceneId
////				status := apis.SceneStatus{
////					Lock:           apis.Lock{Type: apis.MutexLock, IsLocked: true, Ref: action.Status.Scenes[index].Lock.Ref + 1},
////					UpdateTime:     temporaryMap[scene.SceneID].Time,
////					AttachedTask:   temporaryMap[scene.SceneID].AttachedTask,
////					AttachedDevice: temporaryMap[scene.SceneID].AttachedDevice,
////					Property:       map[string]apis.Property{"isOccupied": apis.Property{Name: "isOccupied", Value: "true", Type: apis.BoolType}},
////
////					UpdateMethod:  action.Status.Scenes[index].UpdateMethod,
////					AttachedScene: action.Status.Scenes[index].AttachedScene,
////				}
////				action.Status.Scenes[index] = status
////			}
////		}
////	} else { // TODO:多个设备
////
////	}
////
////	return nil
////}
////
////// RecoverSceneStatus 恢复Scene状态
////func RecoverSceneStatus(runtime *apis.Runtime, action *apis.Action, deviceNumber int, taskId string) error {
////	scenes := runtime.Scenes
////	/* 需要更新的状态有：关联设备 关联任务 时间 关联scene 属性 */
////	// 只有一个设备
////	temporaryMap := GetSceneStatusUpdateMap(runtime, taskId, action)
////	if deviceNumber == 1 {
////		for index, scene := range scenes {
////			switch scene.Type {
////			// object类
////			case apis.ObjectType:
////				// TODO:SceneId Location
////				status := apis.SceneStatus{
////					UpdateTime:     temporaryMap[scene.SceneID].Time,
////					AttachedTask:   "",
////					AttachedDevice: "",
////					Property:       map[string]apis.Property{"location": apis.Property{Name: "location"}},
////					Lock:           apis.Lock{Type: apis.MutexLock, IsLocked: false, Ref: action.Status.Scenes[index].Lock.Ref - 1},
////
////					UpdateMethod:  action.Status.Scenes[index].UpdateMethod,
////					AttachedScene: action.Status.Scenes[index].AttachedScene,
////				}
////				action.Status.Scenes[index] = status
////
////			// position类型
////			case apis.PositionType:
////				// TODO:SceneId
////				status := apis.SceneStatus{
////					Lock:           apis.Lock{Type: apis.MutexLock, IsLocked: true, Ref: action.Status.Scenes[index].Lock.Ref + 1},
////					UpdateTime:     temporaryMap[scene.SceneID].Time,
////					AttachedTask:   temporaryMap[scene.SceneID].AttachedTask,
////					AttachedDevice: temporaryMap[scene.SceneID].AttachedDevice,
////					Property:       map[string]apis.Property{"isOccupied": apis.Property{Name: "isOccupied", Value: "true", Type: apis.BoolType}},
////
////					UpdateMethod:  action.Status.Scenes[index].UpdateMethod,
////					AttachedScene: action.Status.Scenes[index].AttachedScene,
////				}
////				action.Status.Scenes[index] = status
////			}
////		}
////	} else { // TODO:多个设备
////
////	}
////
////	return nil
////}
////
////// GetSceneStatusUpdateMap 负责构建Scene在Update时需要的参数
////func GetSceneStatusUpdateMap(runtime *apis.Runtime, taskId string, action *apis.Action) map[string]TemporaryMap {
////	temporaryMap := make(map[string]TemporaryMap)
////	for _, scene := range runtime.Scenes {
////		temporaryMap[scene.SceneID] = TemporaryMap{
////			// TODO:sceneId
////			// TODO:location
////			AttachedDevice: action.Status.Devices[0].DeviceID,
////			AttachedTask:   taskId,
////			Time:           apis.Time{Time: time.Now()},
////		}
////	}
////	return temporaryMap
////}
